package main

import (
	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/pkg/validators"
)

func Example_respValSets_text() {
	display.Print(&respValSets{
		Data: []*validators.Validator{
			{PubKey: []byte("pubkey1"), Power: 100},
			{PubKey: []byte("pubkey2"), Power: 200},
		},
	}, nil, "text")
	// Output:
	// Current validator set:
	//   0. {pubkey = 7075626b657931, power = 100}
	//   1. {pubkey = 7075626b657932, power = 200}
}

func Example_respValSets_json() {
	display.Print(&respValSets{
		Data: []*validators.Validator{
			{PubKey: []byte("pubkey1"), Power: 100},
			{PubKey: []byte("pubkey2"), Power: 200},
		},
	}, nil, "json")
	// Output:
	// {
	//   "result": [
	//     {
	//       "pubkey": "7075626b657931",
	//       "power": 100
	//     },
	//     {
	//       "pubkey": "7075626b657932",
	//       "power": 200
	//     }
	//   ],
	//   "error": ""
	// }
}

func Example_respValJoinStatus_text() {
	display.Print(&respValJoinStatus{
		Data: &validators.JoinRequest{
			Candidate: []byte("candidate"),
			Power:     100,
			Board:     [][]byte{[]byte("board1"), []byte("board2")},
			Approved:  []bool{true, false},
		}},
		nil, "text")
	// Output:
	// Candidate: 63616e646964617465 (want power 100)
	//  Validator 626f61726431, approved = true
	//  Validator 626f61726432, approved = false
}

func Example_respValJoinStatus_json() {
	display.Print(&respValJoinStatus{
		Data: &validators.JoinRequest{
			Candidate: []byte("candidate"),
			Power:     100,
			Board:     [][]byte{[]byte("board1"), []byte("board2")},
			Approved:  []bool{true, false},
		}},
		nil, "json")
	// Output:
	// {
	//   "result": {
	//     "candidate": "63616e646964617465",
	//     "power": 100,
	//     "Board": [
	//       "626f61726431",
	//       "626f61726432"
	//     ],
	//     "Approved": [
	//       true,
	//       false
	//     ]
	//   },
	//   "error": ""
	// }
}
