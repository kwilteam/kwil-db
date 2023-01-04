package utils

import (
	"kwil/x/crypto"
	"kwil/x/utils"
	"strings"
)

func GenerateSchemaName(owner, name string) string {
	return "x" + crypto.Sha224Str(utils.JoinBytes([]byte(strings.ToLower(name)), []byte(strings.ToLower(owner))))
}
