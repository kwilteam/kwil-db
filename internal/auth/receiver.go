package auth

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	kc "github.com/kwilteam/kwil-db/internal/crypto"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/rand"
	"time"
)

/*
	This file contains the logic for a primary node to accept and authenticate peers
	For the time being, keys will not have an expiration time (will implenent with Redis later)
*/

type Auth struct {
	Authenticator *authenticator
	Client        *authClient
}

type authenticator struct {
	keys map[string]bool
<<<<<<< HEAD
	conf config
=======
	conf *types.Config
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	log  zerolog.Logger
}

// TODO: Authenticator.keys will be ever-growing in memory since deletes don't get reduce the map size
// As long as we switch to redis we should be fine
// If we don't switch to redis, we should routinely copy the map to a new one to prevent memory leaks

<<<<<<< HEAD
func newAuthenticator(c config) *authenticator {
=======
func newAuthenticator(c *types.Config) *authenticator {
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	km := make(map[string]bool)
	logger := log.With().Str("component", "authenticator").Logger()
	return &authenticator{
		keys: km,
		conf: c,
		log:  logger,
	}
}

<<<<<<< HEAD
type config interface {
	IsFriend(string) bool
	GetPeers() []string
}

func NewAuth(c config, a account) *Auth {
=======
func NewAuth(c *types.Config, a account) *Auth {
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	return &Auth{
		Authenticator: newAuthenticator(c),
		Client:        newAuthClient(c, a),
	}
}

/*func (a *Authenticator) ValidateChallenge(c *types.AuthChallengeResponse) (*types.AuthResponse, error) {
}*/

func (a *authenticator) isFriend(s string) bool {
<<<<<<< HEAD
	return a.conf.IsFriend(s)
=======
	return a.conf.Friends[s]
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
}

var ErrAddressNotFriend = errors.New("address is not a friend")
var ErrDifferentNonce = errors.New("nonce is different from challenge")
var ErrSignatureInvalid = errors.New("signature is invalid")

/*
Authenticate will first receive a request containing the address of the client.
If the address is friendly, it will generate a nonce and return it to the client

The client must then send the signed nonce back to the server
If the nonce received by the client is the same as what is stored,
the server will check the signature.

If the ecrecover of the signature matches the address, the server will
store the nonce in the keys map and return a valid response

If the ecrecover of the signature does not match the address, the server will
return an invalid response
*/
func (a *authenticator) Authenticate(c *websocket.Conn) error {
	// Receive first message
	_, msg, err := c.ReadMessage()
	log.Debug().Msgf("received message: %s", msg)
	if err != nil {
		log.Warn().Err(err).Msg("error reading message")
		return err
	}

	// Unmarshal message
	var req types.AuthRequest
	err = json.Unmarshal(msg, &req)
	if err != nil {
		log.Warn().Err(err).Msg("error unmarshaling message")
		return err
	}

	// Check if address is friendly
	if !a.isFriend(req.Address) {
		log.Warn().Err(ErrAddressNotFriend).Msgf("address is not a friend: %s", req.Address)
		return ErrAddressNotFriend
	}

	// Generate nonce
	nonce := generateID(32)

	// Send nonce to client
	cr := types.AuthChallenge{
		Nonce: nonce,
		Valid: true,
	}

	// Marshal response
	bb, err := cr.Bytes()
	if err != nil {
		log.Warn().Err(err).Msg("error marshaling response")
		return err
	}

	// Send response
	err = c.WriteMessage(websocket.TextMessage, bb)
	if err != nil {
		log.Warn().Err(err).Msg("error sending response")
		return err
	}

	// Receive signed nonce
	_, msg, err = c.ReadMessage()
	log.Debug().Msgf("received message: %s", msg)
	if err != nil {
		log.Warn().Err(err).Msg("error reading message")
		return err
	}

	// Unmarshal message
	var csr types.AuthChallengeResponse
	err = json.Unmarshal(msg, &csr)
	if err != nil {
		log.Warn().Err(err).Msg("error unmarshaling message")
		return err
	}

	// Check if nonce is valid
	if csr.Nonce != nonce {
		return ErrDifferentNonce
	}

	// Check if signature is valid
	v, err := kc.CheckSignature(req.Address, csr.Signature, []byte(csr.Nonce))
	if err != nil {
		log.Warn().Err(err).Msg("error checking signature")
		return err
	}
	if !v {
		log.Warn().Err(ErrSignatureInvalid).Msg("signature is invalid")
		return ErrSignatureInvalid
	}

	// Store nonce in keys map
	a.keys[csr.Nonce] = true

	// Send response
	ar := types.AuthResponse{
		Nonce: csr.Nonce,
		Valid: true,
	}

	// Marshal response
	bb, err = ar.Bytes()
	if err != nil {
		log.Warn().Err(err).Msg("error marshaling response")
		return err
	}

	// Send response
	err = c.WriteMessage(websocket.TextMessage, bb)
	if err != nil {
		log.Warn().Err(err).Msg("error sending response")
		return err
	}

	return nil
}

// generateID generates a random ID for the request
func generateID(l uint8) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
