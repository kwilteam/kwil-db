package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	update = flag.Bool("update", false, "update the golden files of this test")
)

var testDataDir = "testdata"

func goldenValue(t *testing.T, goldenFile string, actual string, update bool) string {
	t.Helper()
	goldenPath := filepath.Join(".", testDataDir, goldenFile+"_expect.json")

	f, err := os.OpenFile(goldenPath, os.O_RDWR|os.O_CREATE, 0666)
	require.NoError(t, err)
	defer f.Close()

	if update {
		_, err := f.WriteString(actual)
		if err != nil {
			t.Fatalf("Error writing to file %s: %s", goldenPath, err)
		}

		return actual
	}

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Error opening file %s: %s", goldenPath, err)
	}
	return string(content)
}

func Test_parseEstimateCostResponse(t *testing.T) {
	tests := []struct {
		name string
		resp []byte
		want *big.Int
	}{
		{
			name: "ok",
			resp: []byte(`{"price":"100"}`),
			want: big.NewInt(100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseEstimateCostResponse(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseBroadcastResponse(t *testing.T) {
	tests := []struct {
		name       string
		resp       []byte
		wantBase64 string
	}{
		{
			name:       "ok",
			resp:       []byte(`{"tx_hash":"FawI7Hzfc3lC9zqyWvO6xeXAsCGjsI99d/cjXShFoXU="}`),
			wantBase64: "FawI7Hzfc3lC9zqyWvO6xeXAsCGjsI99d/cjXShFoXU=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseBroadcastResponse(&response)
			assert.NoError(t, err)

			want, _ := base64.StdEncoding.DecodeString(tt.wantBase64)
			assert.Equal(t, want, got)
		})
	}
}

func Test_parseGetAccountResponse(t *testing.T) {
	// got lazy so no need to construct the whole parsed *types.Account
	// `go test -run Test_parseGetAccountResponse . -update` to update _expect.json
	tests := []struct {
		name       string
		target     string
		statusCode int
	}{
		{
			name:       "ok but not exist",
			target:     "get_account_not_exist",
			statusCode: http.StatusOK,
		},
		{
			name:       "ok",
			target:     "get_account_ok",
			statusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPath := filepath.Join(".", testDataDir, tt.target+"_response.json")
			data, err := os.ReadFile(jsonPath)
			require.NoError(t, err)

			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(data)),
			}

			gotObj, err := parseGetAccountResponse(&response)
			if tt.statusCode != http.StatusOK {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			got, err := json.MarshalIndent(gotObj, "", "  ")
			assert.NoError(t, err)

			want := goldenValue(t, tt.target, string(got), *update)
			assert.Equal(t, want, string(got))
		})
	}
}

func Test_parseTxQueryResponse(t *testing.T) {
	// NOTE: probably should construct *transactions.TcTxQueryResponse
	// ./testdata/NAME_response.json
	// ./testdata/NAME_expect.json
	// the difference between _response.json and _expect.json is that
	// int64 is string in _response.json
	// `go test -run Test_parseTxQueryResponse . -update` to update _expect.json

	tests := []struct {
		name       string
		target     string
		statusCode int
		errMsg     string // this sucks
	}{
		{
			name:       "ok",
			target:     "tx_query_ok",
			statusCode: http.StatusOK,
		},
		{
			name:       "tx not found",
			target:     "tx_query_not_found",
			statusCode: http.StatusNotFound,
			errMsg:     "transaction not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPath := filepath.Join(".", testDataDir, tt.target+"_response.json")
			data, err := os.ReadFile(jsonPath)
			require.NoError(t, err)

			response := http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewReader(data)),
			}

			gotObj, err := parseTxQueryResponse(&response)
			if tt.statusCode != http.StatusOK {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)

			got, err := json.MarshalIndent(gotObj, "", "  ")
			assert.NoError(t, err)

			want := goldenValue(t, tt.target, string(got), *update)
			assert.Equal(t, want, string(got))
		})
	}
}

func Test_parseListDatabasesResponse(t *testing.T) {
	tests := []struct {
		name string
		resp []byte
		want []string
	}{
		{
			name: "ok",
			resp: []byte(`{
    "databases": [
        "testdb"
    ]
}`),
			want: []string{"testdb"},
		},
		{
			name: "ok but no databases",
			resp: []byte(`{
    "databases": []
}`),
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseListDatabasesResponse(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseActionCallResponse(t *testing.T) {
	tests := []struct {
		name string
		resp []byte
		want []map[string]any
	}{
		{
			name: "ok",
			// base64 of [{"owner only":"owner only"}]
			resp: []byte(`{"result":"W3siJ293bmVyIG9ubHknIjoib3duZXIgb25seSJ9XQ=="}`),
			want: []map[string]any{
				{
					"'owner only'": "owner only",
				},
			},
		},
		{
			name: "data",
			resp: []byte(`{"result":"W3siYWdlIjozMywiaWQiOjIsInVzZXJuYW1lIjoic2F0b3NoaSJ9XQ=="}`),
			want: []map[string]any{
				{
					"age":      float64(33), // json unmarshal to float64
					"id":       float64(2),
					"username": "satoshi",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseActionCallResponse(&response)
			assert.NoError(t, err)

			assert.EqualValues(t, tt.want, got)
		})
	}
}

func Test_parseDBQueryResponse(t *testing.T) {
	tests := []struct {
		name string
		resp []byte
		want []map[string]any
	}{
		{
			name: "ok",
			// base64 of [{"owner only":"owner only"}]
			resp: []byte(`{"result":"W3siJ293bmVyIG9ubHknIjoib3duZXIgb25seSJ9XQ=="}`),
			want: []map[string]any{
				{
					"'owner only'": "owner only",
				},
			},
		},
		{
			name: "data",
			resp: []byte(`{"result":"W3siYWdlIjozMywiaWQiOjIsInVzZXJuYW1lIjoic2F0b3NoaSJ9XQ=="}`),
			want: []map[string]any{
				{
					"age":      float64(33), // json unmarshal to float64
					"id":       float64(2),
					"username": "satoshi",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseDBQueryResponse(&response)
			assert.NoError(t, err)

			assert.EqualValues(t, tt.want, got)
		})
	}
}

func Test_parseGetSchemaResponse(t *testing.T) {
	// go test -run Test_parseGetSchemaResponse . -update
	tests := []struct {
		name       string
		target     string
		statusCode int
		errMsg     string // this sucks
	}{
		{
			name:       "ok",
			target:     "get_schema_ok",
			statusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPath := filepath.Join(".", testDataDir, tt.target+"_response.json")
			data, err := os.ReadFile(jsonPath)
			require.NoError(t, err)

			response := http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewReader(data)),
			}

			gotObj, err := parseGetSchemaResponse(&response)
			if tt.statusCode != http.StatusOK {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)

			got, err := json.MarshalIndent(gotObj, "", "  ")
			assert.NoError(t, err)

			want := goldenValue(t, tt.target, string(got), *update)
			assert.Equal(t, want, string(got))
		})
	}
}

func Test_parseValidatorJoinStatusResponse(t *testing.T) {
	pbk1Str := "BIEr70T257KhnAsBwtyl5Uuhk1oYkP/cuTq9DFNLIJwh5PYXaCP+9JP3ta+qRW8x1SkzY9j4AcVA68wGGBKJDLo="
	pbk2Str := "BIEr70T257KhnAsBwtyl5Uuhk1oYkP/cuTq9DFNLIJwh5PYXaCP+9JP3ta+qRW8x1SkzY9j4AcVA68wGGBKJDL0="
	pubKey1, _ := base64.StdEncoding.DecodeString(pbk1Str)
	pubKey2, _ := base64.StdEncoding.DecodeString(pbk2Str)

	tests := []struct {
		name string
		resp []byte
		want *types.JoinRequest
	}{
		{
			name: "ok",
			resp: []byte(fmt.Sprintf(`{
   "approved_validators":[
      "%s"
   ],
   "pending_validators":[
      "%s"
   ],
   "power": "10"
}`, pbk1Str, pbk2Str)),
			want: &types.JoinRequest{
				Power: 10,
				Board: [][]byte{
					pubKey1,
					pubKey2,
				},
				Approved: []bool{
					true,
					false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseValidatorJoinStatusResponse(&response, nil)
			assert.NoError(t, err)

			assert.EqualValues(t, tt.want, got)
		})
	}
}

func Test_parseCurrentValidatorsResponse(t *testing.T) {
	pbk1Str := "BIEr70T257KhnAsBwtyl5Uuhk1oYkP/cuTq9DFNLIJwh5PYXaCP+9JP3ta+qRW8x1SkzY9j4AcVA68wGGBKJDLo="
	pbk2Str := "BIEr70T257KhnAsBwtyl5Uuhk1oYkP/cuTq9DFNLIJwh5PYXaCP+9JP3ta+qRW8x1SkzY9j4AcVA68wGGBKJDL0="
	pubKey1, _ := base64.StdEncoding.DecodeString(pbk1Str)
	pubKey2, _ := base64.StdEncoding.DecodeString(pbk2Str)

	tests := []struct {
		name string
		resp []byte
		want []*types.Validator
	}{
		{
			name: "ok",
			resp: []byte(fmt.Sprintf(`{
  "validators": [
    {
      "pubkey": "%s",
      "power": "10"
    },
    {
      "pubkey": "%s",
      "power": "10"
    }
  ]
}
`, pbk1Str, pbk2Str)),
			want: []*types.Validator{
				{
					PubKey: pubKey1,
					Power:  10,
				},
				{
					PubKey: pubKey2,
					Power:  10,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseCurrentValidatorsResponse(&response)
			assert.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseChainInfoResponse(t *testing.T) {
	tests := []struct {
		name string
		resp []byte
		want *types.ChainInfo
	}{
		{
			name: "ok",
			resp: []byte(`{
    "chain_id": "kwil-chain-XHAPEDGA",
    "height": "375",
    "hash": "bfacc4bfc60620dec2456f3e66a369ce60ceb53402dd600163c68e1bbffb87e8"
}`),
			want: &types.ChainInfo{
				ChainID:     "kwil-chain-XHAPEDGA",
				BlockHeight: 375,
				BlockHash:   "bfacc4bfc60620dec2456f3e66a369ce60ceb53402dd600163c68e1bbffb87e8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseChainInfoResponse(&response)
			assert.NoError(t, err)

			assert.EqualValues(t, tt.want, got)
		})
	}
}

func Test_parseVerifySignatureRespoonse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		resp       []byte
		wantErr    bool
		want       bool
	}{
		{
			name:       "other error",
			statusCode: http.StatusInternalServerError,
			resp: []byte(`{
  "code": 5,
  "message": "something happen",
  "details": []
}`),
			want:    false,
			wantErr: true,
		},
		{
			name:       "ok",
			statusCode: http.StatusOK,
			resp:       []byte(`{"valid":true, "error": ""}`),
			want:       true,
		},
		{
			name:       "invalid",
			statusCode: http.StatusOK,
			resp:       []byte(`{"valid":false, "error": "some reason"}`),
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewReader(tt.resp)),
			}

			got, err := parseVerifySignatureResponse(&response)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
