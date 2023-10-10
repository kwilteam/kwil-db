package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/validators"
)

// respValSets represent current validator set in cli
type respValSets struct {
	Data []*validators.Validator
}

type valInfo struct {
	PubKey string `json:"pubkey"`
	Power  int64  `json:"power"`
}

func (r *respValSets) MarshalJSON() ([]byte, error) {
	valInfos := make([]valInfo, len(r.Data))
	for i, v := range r.Data {
		valInfos[i] = valInfo{
			PubKey: fmt.Sprintf("%x", v.PubKey),
			Power:  v.Power,
		}
	}

	return json.Marshal(valInfos)
}

func (r *respValSets) MarshalText() ([]byte, error) {
	var msg bytes.Buffer
	msg.WriteString("Current validator set:\n")
	for i, v := range r.Data {
		msg.WriteString(fmt.Sprintf("% 3d. %s\n", i, v))
	}

	return msg.Bytes(), nil
}

// respValJoinStatus represent the status of a validator join request in cli
type respValJoinStatus struct {
	Data *validators.JoinRequest
}

// respValJoinRequest is customized json format for respValJoinStatus
type respValJoinRequest struct {
	Candidate string `json:"candidate"`
	Power     int64  `json:"power"`
	Board     []string
	Approved  []bool
}

func (r *respValJoinStatus) MarshalJSON() ([]byte, error) {
	joinReq := &respValJoinRequest{
		Candidate: fmt.Sprintf("%x", r.Data.Candidate),
		Power:     r.Data.Power,
		Board:     make([]string, len(r.Data.Board)),
		Approved:  r.Data.Approved,
	}
	for i := range r.Data.Board {
		joinReq.Board[i] = fmt.Sprintf("%x", r.Data.Board[i])
	}

	return json.Marshal(joinReq)
}

func (r *respValJoinStatus) MarshalText() ([]byte, error) {
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("Candidate: %x (want power %d)\n",
		r.Data.Candidate, r.Data.Power))
	for i := range r.Data.Board {
		msg.WriteString(fmt.Sprintf(" Validator %x, approved = %v\n",
			r.Data.Board[i], r.Data.Approved[i]))
	}

	return msg.Bytes(), nil
}
