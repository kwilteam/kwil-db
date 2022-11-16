package schema

import (
	"bytes"
	"io"
	"ksl"
	"ksl/backend"
	"ksl/pdb"
)

type KwilSchema struct {
	Db          *pdb.Db
	Diagnostics ksl.Diagnostics
	Backend     backend.Connector
}

func New(db *pdb.Db, diags ksl.Diagnostics) *KwilSchema {
	backend := backend.Get(db.Config.BackendName())
	return &KwilSchema{Db: db, Diagnostics: pdb.NewValidationContext(db, diags).Validate(db), Backend: backend}
}

func (s *KwilSchema) HasErrors() bool { return s.Diagnostics.HasErrors() }

func (s *KwilSchema) WriteDiagnostics(w io.Writer, console bool) {
	wr := ksl.NewDiagnosticTextWriter(w, s.Db.Ast.Sources(), 120, console)
	wr.WriteDiagnostics(s.Diagnostics)
}

func (s *KwilSchema) Data() []byte {
	var buf bytes.Buffer
	for _, file := range s.Db.Ast.Files {
		buf.Write(file.Contents)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}
