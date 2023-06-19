package dataset2

import (
	"fmt"
)

// a procedure is a collection of operations that can be executed as a single unit
// it is atomic, and has local variables
type Procedure struct {
	Name       string           `json:"name"`
	Parameters []string         `json:"args"`
	Scoping    ProcedureScoping `json:"scoping"`
	Body       []*OpCodeExecution
}

type ProcedureScoping uint8

const (
	ProcedureScopingPublic ProcedureScoping = iota
	ProcedureScopingPrivate
)

type Operation1 interface {
	// args is the list of required arguments for the operation.
	args() []string
	// returns is the list of variables names assigned by the operation.
	returns() []string

	// prepare prepares the operation for evaluation.
	prepare(*Dataset) (operation, error)

	Type() OperationType
}

type OperationType uint8

func (o OperationType) Byte() byte {
	return byte(o)
}

const (
	OperationTypeDML OperationType = iota
	OperationTypeExtensionMethod
	OperationTypeProcedureCall
)

type DMLOperation struct {
	Statement string `json:"statement"`
}

func (o *DMLOperation) args() []string {
	return []string{}
}

func (o *DMLOperation) returns() []string {
	return []string{}
}

func (o *DMLOperation) prepare(d *Dataset) (operation, error) {
	preparedStmt, err := d.db.Prepare(o.Statement)
	if err != nil {
		return nil, err
	}

	return &dmlStatement{
		stmt: preparedStmt,
	}, nil
}

func (o *DMLOperation) Type() OperationType {
	return OperationTypeDML
}

type ExtensionMethodOperation struct {
	Extension string   `json:"extension"`
	Method    string   `json:"method"`
	Args      []string `json:"args"`
	Returns   []string `json:"returns"`
}

func (o *ExtensionMethodOperation) args() []string {
	return o.Args
}

func (o *ExtensionMethodOperation) returns() []string {
	return o.Returns
}

func (o *ExtensionMethodOperation) prepare(d *Dataset) (operation, error) {
	extensions, ok := d.extensions[o.Extension]
	if !ok {
		return nil, fmt.Errorf("unknown extension '%s'", o.Extension)
	}

	return &extensionMethod{
		extensions:   extensions,
		method:       o.Method,
		requiredArgs: o.Args,
		returns:      o.Returns,
	}, nil
}

func (o *ExtensionMethodOperation) Type() OperationType {
	return OperationTypeExtensionMethod
}

// A ProcedureCallOperation calls a procedure from within another procedure.
type ProcedureCallOperation struct {
	// The name of the procedure to call.
	Procedure string `json:"procedure"`

	// The arguments to pass to the procedure.  Should be valid variable names, starting with $ or @.
	Args []string `json:"args"`
}

func (o *ProcedureCallOperation) args() []string {
	return o.Args
}

func (o *ProcedureCallOperation) returns() []string {
	return []string{}
}

func (o *ProcedureCallOperation) prepare(d *Dataset) (operation, error) {
	procedure, ok := d.procedures[o.Procedure]
	if !ok {
		return nil, fmt.Errorf("unknown procedure '%s'", o.Procedure)
	}

	return &procedureExecution{
		procedure: procedure,
		args:      o.Args,
	}, nil
}

func (o *ProcedureCallOperation) Type() OperationType {
	return OperationTypeProcedureCall
}
