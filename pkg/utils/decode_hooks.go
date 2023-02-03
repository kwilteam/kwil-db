package utils

import (
	"crypto/ecdsa"
	"fmt"
	ec "github.com/ethereum/go-ethereum/crypto"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

func StringPrivateKeyHookFunc() mapstructure.DecodeHookFuncType {
	return func(f, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String && t == reflect.TypeOf(&ecdsa.PrivateKey{}) {

			d := data.(string)
			if d == "" {
				return nil, nil
			}

			privateKey, err := ec.HexToECDSA(data.(string))
			if err != nil {
				return nil, fmt.Errorf("error parsing private key: %v", err)
			}
			return privateKey, nil
		}
		return data, nil
	}
}
