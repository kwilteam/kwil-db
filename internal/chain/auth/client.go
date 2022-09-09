package auth

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
	types "github.com/kwilteam/kwil-db/pkg/types/chain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type account interface {
	Sign(data []byte) (string, error)
	GetAddress() string
}

type authClient struct {
	keys map[string]string // maps ip to nonce
	acc  account
	log  zerolog.Logger
	conf config
}

func newAuthClient(c config, a account) *authClient {
	logger := log.With().Str("component", "auth_client").Logger()
	return &authClient{
		keys: make(map[string]string),
		acc:  a,
		log:  logger,
		conf: c,
	}
}

var ErrConnectionDropped = errors.New("connection dropped")
var ErrAuthRefused = errors.New("authentication refused")

/*
	RequestAuth will create a websocket connection with the ip at the /peer-auth endpoint

	Once established, it will send a AuthRequest message to the peer

	The peer will respond with a AuthChallenge message

	The client will then sign the nonce and send it back to the peer

	The peer will then verify the signature and respond with a AuthResponse message

	If the response is valid, the client will store the nonce in the keys map
	If invalid, the client will return false

	If the connection is closed at any point, the client will return false
*/

// AuthAll will try to authenticate with all peers in the peer list
func (ac *authClient) AuthAll() {
	for _, p := range ac.conf.GetPeers() {
		ok, err := ac.RequestAuth(p)
		if err != nil || !ok {
			log.Warn().Err(err).Msgf("failed to authenticate with peer %s", p)
		}
	}
}

func (ac *authClient) RequestAuth(ip string) (bool, error) {
	// Create a new websocket connection
	c, _, err := websocket.DefaultDialer.Dial("ws://"+ip+"/peer-auth", nil)
	if err != nil {
		log.Warn().Err(err).Msg("failed to connect to peer")
		return false, err
	}
	log.Debug().Msgf("connected to peer at %s", ip)

	// Send AuthRequest
	ar := &types.AuthRequest{
		Address: ac.acc.GetAddress(),
	}

	// Convert ar to bytes
	b, err := ar.Bytes()
	if err != nil {
		log.Warn().Err(err).Msg("failed to convert AuthRequest to bytes")
		return false, err
	}

	// Send AuthRequest
	err = c.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Warn().Err(err).Msg("failed to send AuthRequest")
		return false, err
	}

	// Receive AuthChallenge
	_, b, err = c.ReadMessage()
	if err != nil {
		log.Warn().Err(err).Msg("failed to receive AuthChallenge")
		return false, err
	}
	//return true, nil
	// Unmarshal AuthChallenge
	acm := &types.AuthChallenge{}
	err = json.Unmarshal(b, acm)
	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal AuthChallenge")
		return false, err
	}

	// Sign nonce
	sig, err := ac.acc.Sign([]byte(acm.Nonce))
	if err != nil {
		log.Warn().Err(err).Msg("failed to sign nonce")
		return false, err
	}

	// Send AuthChallengeResponse
	acr := &types.AuthChallengeResponse{
		Signature: sig,
		Nonce:     acm.Nonce,
	}

	// Convert acr to bytes
	b, err = acr.Bytes()
	if err != nil {
		log.Warn().Err(err).Msg("failed to convert AuthChallengeResponse to bytes")
		return false, err
	}

	// Send AuthChallengeResponse
	err = c.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Warn().Err(err).Msg("failed to send AuthChallengeResponse")

		return false, err
	}

	// Receive AuthResponse
	_, b, err = c.ReadMessage()
	if err != nil {
		log.Warn().Err(err).Msg("failed to receive AuthResponse")
		return false, err
	}

	// Unmarshal AuthResponse
	arm := &types.AuthResponse{}
	err = json.Unmarshal(b, arm)
	if err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal AuthResponse")
		return false, err
	}

	// Check if nonce is same
	if arm.Nonce != acm.Nonce {
		log.Warn().Msg("nonce does not match")
		return false, ErrDifferentNonce
	}

	// Check if response is valid
	if !arm.Valid {
		log.Warn().Msg("response is invalid")
		return false, ErrAuthRefused
	}

	// Store nonce in keys map
	ac.keys[ip] = acm.Nonce

	return true, nil
}
