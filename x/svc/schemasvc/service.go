package schemasvc

import (
	schemapb "kwil/x/proto/schemasvc"
)

type service struct {
	schemapb.UnimplementedSchemaServiceServer
}

func New() schemapb.SchemaServiceServer {
	return &service{}
}
