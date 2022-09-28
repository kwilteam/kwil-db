package service

import (
	"strings"

	proto "kwil/x/api/v0"
	"kwil/x/chain/crypto"
)

func createDatabaseID(owner, name, fee string) string {
	sb := strings.Builder{}
	sb.WriteString(owner)
	sb.WriteString(name)
	sb.WriteString(fee)
	return string(crypto.Sha384([]byte(sb.String())))
}

func updateDatabaseID(m *proto.UpdateDatabaseRequest) string {
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
