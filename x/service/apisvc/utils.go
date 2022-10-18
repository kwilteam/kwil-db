package apisvc

import (
	"strings"

	"kwil/x/crypto"
	"kwil/x/proto/apipb"
)

func createDatabaseID(owner, name, fee string) string {
	sb := strings.Builder{}
	sb.WriteString(owner)
	sb.WriteString(name)
	sb.WriteString(fee)
	return string(crypto.Sha384([]byte(sb.String())))
}

func createFundsReturnID(amount, nonce, address string) string {
	sb := strings.Builder{}
	sb.WriteString(amount)
	sb.WriteString(nonce)
	sb.WriteString(address)
	return string(crypto.Sha384([]byte(sb.String())))
}

func updateDatabaseID(m *apipb.UpdateDatabaseRequest) string {
	sb := strings.Builder{}
	sb.WriteString(m.Owner)
	sb.WriteString(m.Name)
	sb.WriteString(m.Fee)
	sb.WriteByte(byte(m.Operation))
	sb.WriteByte(byte(m.Crud))
	sb.WriteString(m.Instructions)
	sb.WriteString(m.From)
	sb.WriteString(m.Nonce)

	return string(crypto.Sha384([]byte(sb.String())))
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
