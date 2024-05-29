package usersvc

//go:generate go run genopenrpcspec.go

import (
	"reflect"

	"github.com/kwilteam/kwil-db/core/rpc/json/openrpc"
)

var (
	SpecInfo = openrpc.Info{
		Title:       "Kwil DB user service",
		Description: `The JSON-RPC user service for Kwil DB.`,
		License: &openrpc.License{ // the spec file's license
			Name: "CC0-1.0",
			URL:  "https://creativecommons.org/publicdomain/zero/1.0/legalcode",
		},
		Version: "0.1.0",
	}
)

func OpenRPCSpec() openrpc.Spec {
	svc := &Service{}
	methodDefs := make(map[string]*openrpc.MethodDefinition)
	for method, def := range svc.Methods() {
		methodDefs[string(method)] = &openrpc.MethodDefinition{
			Description:  def.Desc,
			RequestType:  def.ReqType,
			ResponseType: def.RespType,
			RespTypeDesc: def.RespDesc,
		}
	}
	knownSchemas := make(map[reflect.Type]openrpc.Schema)
	methods := openrpc.InventoryAPI(methodDefs, knownSchemas)
	schemas := make(map[string]openrpc.Schema)
	for _, schema := range knownSchemas {
		schemas[schema.Name()] = schema
	}
	return openrpc.Spec{
		OpenRPC: "1.2.4",
		Info:    SpecInfo,
		Methods: methods,
		Components: openrpc.Components{
			Schemas: schemas,
		},
	}
}
