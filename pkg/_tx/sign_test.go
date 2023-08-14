package tx_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

const defaultPrivateKey = "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"

func Test_Sign(t *testing.T) {
	type testCase struct {
		name          string
		payload       tx.Serializable
		signer        string
		checkedSigner string
		wantErr       bool
	}

	testCases := []testCase{
		{
			name: "call payload",
			payload: &tx.CallActionPayload{
				Action: "get_post",
				DBID:   "dbid",
				Params: map[string]any{
					"$id": 1,
				},
			},
			signer:        defaultPrivateKey,
			checkedSigner: defaultPrivateKey,
			wantErr:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signer, err := crypto.ECDSAFromHex(tc.signer)
			if err != nil {
				t.Fatalf("failed to create private key: %s", err.Error())
			}

			checkSigner, err := crypto.ECDSAFromHex(tc.checkedSigner)
			if err != nil {
				t.Fatalf("failed to create private key: %s", err.Error())
			}

			signedMsg, err := tx.CreateSignedMessage(tc.payload, signer)
			if err != nil {
				t.Fatalf("failed to sign payload: %s", err.Error())
			}

			signedMsg.Sender = crypto.AddressFromPrivateKey(checkSigner)

			err = signedMsg.Verify()
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %s", err.Error())
				}
			}
		})

	}
}
