package apisvc

import (
	"strings"

	"kwil/x/chain/crypto"
	"kwil/x/proto/apipb"
)

func createDatabaseID(owner, name, fee string) string {
	sb := strings.Builder{}
	sb.WriteString(owner)
	sb.WriteString(name)
	sb.WriteString(fee)
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
