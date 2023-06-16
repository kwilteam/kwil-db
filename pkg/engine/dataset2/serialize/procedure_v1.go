package serialize

import (
	"github.com/kwilteam/kwil-db/pkg/engine/dataset2"
)

// procedurev1 is the v1 serialization format for procedures.
// it was previously known as "proc".
type procedure_v1 struct {
	Name       string   `json:"name"`
	Inputs     []string `json:"inputs"`
	Public     bool     `json:"public"`
	Statements []string `json:"statements"`
}

func procedureV1ToV2(proc *procedure_v1) (*dataset2.Procedure, error) {
	var scoping dataset2.ProcedureScoping
	if proc.Public {
		scoping = dataset2.ProcedureScopingPublic
	} else {
		scoping = dataset2.ProcedureScopingPrivate
	}

	var body []dataset2.Operation
	for _, stmt := range proc.Statements {
		body = append(body, procedure_v1_stmt_to_operation(stmt))
	}

	return &dataset2.Procedure{
		Name:    proc.Name,
		Args:    proc.Inputs,
		Scoping: scoping,
		Body:    body,
	}, nil
}

func procedure_v1_stmt_to_operation(stmt string) dataset2.Operation {
	return &dataset2.DMLOperation{
		Statement: stmt,
	}
}
