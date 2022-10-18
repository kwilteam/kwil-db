package schemasvc

import (
	"kwil/x/proto/schemapb"
)

type service struct {
	schemapb.UnimplementedSchemaServiceServer
}

func New() schemapb.SchemaServiceServer {
	return &service{}
}
