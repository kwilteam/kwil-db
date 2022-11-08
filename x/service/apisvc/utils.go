package apisvc

import (
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

func planID(m *apipb.PlanRequest) string {
	sb := strings.Builder{}
	sb.WriteString(m.DbName)
	sb.WriteString(m.Owner)
	sb.WriteString(m.From)
	sb.Write(m.Schema)

	return string(crypto.Sha384([]byte(sb.String())))
}

func applyID(m *apipb.ApplyRequest) string {
	sb := strings.Builder{}
	sb.WriteString(m.PlanId)
	sb.WriteString(m.Fee)
	sb.WriteString(m.Nonce)
	sb.WriteString(m.From)

	return string(crypto.Sha384([]byte(sb.String())))
}
