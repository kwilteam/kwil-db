package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/types"
)

type respChainInfo struct {
	Info *types.ChainInfo
}

func (r *respChainInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Info)
}

func (r *respChainInfo) MarshalText() ([]byte, error) {
	msg := fmt.Sprintf(`Chain ID: %s
Height: %d
Hash: %s
`,
		r.Info.ChainID,
		r.Info.BlockHeight,
		r.Info.BlockHash,
	)

	return []byte(msg), nil
}

// respKwilCliConfig is used to represent a kwil-cli config in cli
type respKwilCliConfig struct {
	cfg *config.KwilCliConfig
}

func (r *respKwilCliConfig) MarshalJSON() ([]byte, error) {
	cfg := r.cfg.ToPersistedConfig()
	cfg.PrivateKey = "***"
	return json.Marshal(cfg)
}

func (r *respKwilCliConfig) MarshalText() ([]byte, error) {
	var msg bytes.Buffer
	cfg := r.cfg.ToPersistedConfig()
	cfg.PrivateKey = "***"

	v := reflect.ValueOf(cfg)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		msg.WriteString(fmt.Sprintf("%s: %v\n", field.Name, fieldValue))
	}

	return msg.Bytes(), nil
}
