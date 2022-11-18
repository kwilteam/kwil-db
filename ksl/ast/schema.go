package ast

import (
	"bytes"
	"io"
	"ksl"
	"ksl/ast/pdb"
	"ksl/backend"
)

type SchemaAst struct {
	Db          *pdb.Db
	Diagnostics ksl.Diagnostics
	Backend     backend.Connector
}

func New(db *pdb.Db, diags ksl.Diagnostics) *SchemaAst {
	backend := backend.Get(db.Config.BackendName())
	return &SchemaAst{Db: db, Diagnostics: pdb.NewValidationContext(db, diags).Validate(db), Backend: backend}
}

func (s *SchemaAst) HasErrors() bool { return s.Diagnostics.HasErrors() }

func (s *SchemaAst) WriteDiagnostics(w io.Writer, console bool) {
	wr := ksl.NewDiagnosticTextWriter(w, s.Db.Ast.Sources(), 120, console)
	wr.WriteDiagnostics(s.Diagnostics)
}

func (s *SchemaAst) Data() []byte {
	var buf bytes.Buffer
	for _, file := range s.Db.Ast.Files {
		buf.Write(file.Contents)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}
