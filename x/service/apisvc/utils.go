package apisvc

import (
	"encoding/json"
	"math/big"
	"strings"

	"kwil/x/crypto"
	"kwil/x/proto/apipb"
)

func createFundsReturnID(amount, nonce, address string) string {
	sb := strings.Builder{}
	sb.WriteString(amount)
	sb.WriteString(nonce)
	sb.WriteString(address)
	return crypto.Sha384Str([]byte(sb.String()))
}

func deployID(m *apipb.DeploySchemaRequest) string {
	sb := strings.Builder{}
	sb.Write(m.Data)
	sb.WriteString(m.Fee)
	sb.WriteString(m.Nonce)
	sb.WriteString(m.From)
	return crypto.Sha384Str([]byte(sb.String()))
}

// query, inputs, fee, nonce hash
func cudID(m *apipb.CUDRequest) string {
	sb := strings.Builder{}
	bts, err := json.Marshal(m.Inputs)
	if err != nil {
		return ""
	}
	sb.WriteString(m.Query)
	sb.Write(bts)
	sb.WriteString(m.Fee)
	sb.WriteString(m.Nonce)

	return crypto.Sha384Str([]byte(sb.String()))
}

func planID(m *apipb.PlanRequest) string {
	sb := strings.Builder{}
	sb.WriteString(m.DbName)
	sb.WriteString(m.Owner)
	sb.WriteString(m.From)
	sb.Write(m.Schema)

	return crypto.Sha384Str([]byte(sb.String()))
}

func applyID(m *apipb.ApplyRequest) string {
	sb := strings.Builder{}
	sb.WriteString(m.PlanId)
	sb.WriteString(m.Fee)
	sb.WriteString(m.Nonce)
	sb.WriteString(m.From)

	return crypto.Sha384Str([]byte(sb.String()))
}

func parseBigInt(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 10)
}
