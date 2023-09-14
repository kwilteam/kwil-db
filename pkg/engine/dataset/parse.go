package dataset

import (
	"context"
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/engine/dataset/actparser"
	"github.com/kwilteam/kwil-db/pkg/engine/execution"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"go.uber.org/zap"

	"encoding/binary"
	"math/rand"
	"time"
)

// randomIdentifier is a hash that can be used and reset.  It is primatily used for generating
// random names for prepared statements.  it is pseudoprandom, and does not have to be
// deterministic.
// When used, it should update to the most recent hash.
type randomIdentifier struct {
	mu   sync.Mutex
	hash string
}

func (r *randomIdentifier) getRandomHash() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	newHash := crypto.Sha224Hex([]byte(r.hash))
	r.hash = newHash

	return newHash
}

func (r *randomIdentifier) getRandomIdent() string {
	return "!" + r.getRandomHash()
}

var randomHash randomIdentifier

func init() {
	// Initialize the random number generator with a seed based on the current time
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random number between 0 and 99
	randomNumber := rand.Intn(100)

	// Convert the random number to bytes
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(randomNumber))

	randomHash = randomIdentifier{
		mu:   sync.Mutex{},
		hash: crypto.Sha224Hex(buf),
	}
}

// TODO: refactor this.
func (d *Dataset) getEngineOpts(ctx context.Context, exts map[string]Initializer) (*execution.EngineOpts, error) {
	procedures, err := d.db.ListProcedures(ctx)
	if err != nil {
		return nil, err
	}

	extensions, err := d.db.ListExtensions(ctx)
	if err != nil {
		return nil, err
	}

	engineProceduresMap := make(map[string]*execution.Procedure)
	loaders := []*execution.InstructionExecution{}
	for _, procedure := range procedures {
		engineProcedure, loaderInstructions, err := parseAction(procedure)
		if err != nil {
			return nil, err
		}

		engineProceduresMap[procedure.Name] = engineProcedure
		loaders = append(loaders, loaderInstructions...)
	}

	for _, extension := range extensions {
		var loaderInstructions []*execution.InstructionExecution

		if _, ok := exts[extension.Name]; !ok {
			d.log.Warn("extension used in dataset not found", zap.String("extension", extension.Name), zap.String("DBID", d.DBID()))

			if d.allowMissingExtensions {
				continue
			}

			return nil, fmt.Errorf("%w: %s", ErrExtensionNotFound, extension.Name)

		} else {
			loaderInstructions, err = getExtensionLoader(extension)
			if err != nil {
				return nil, err
			}
		}

		loaders = append(loaders, loaderInstructions...)
	}

	extensionMap := make(map[string]execution.Initializer)
	for name, initializer := range exts {
		extensionMap[name] = initializerWrapper{initializer}
	}

	return &execution.EngineOpts{
		Extensions: extensionMap,
		Procedures: engineProceduresMap,
		LoadCmd:    loaders,
	}, nil
}

func getExtensionLoader(extension *types.Extension) ([]*execution.InstructionExecution, error) {
	var setters []*execution.InstructionExecution
	configMap := make(map[string]string)
	for key, val := range extension.Initialization {
		if isIdent(val) {
			configMap[key] = val
			continue
		}

		loadOp, name := newSetter(val)

		configMap[key] = name
		setters = append(setters, loadOp)
	}

	initializeOp := &execution.InstructionExecution{
		Instruction: execution.OpExtensionInitialize,
		Args: []any{
			extension.Name,
			extension.Alias,
			configMap,
		},
	}

	return append(setters, initializeOp), nil
}

func parseAction(procedure *types.Procedure) (engineProcedure *execution.Procedure, loaderInstructions []*execution.InstructionExecution, err error) {
	var procedureInstructions []*execution.InstructionExecution
	loaderInstructions = []*execution.InstructionExecution{}

	mut := procedure.IsMutative()
	for _, stmt := range procedure.Statements {
		parsedStmt, err := actparser.Parse(stmt)
		if err != nil {
			return nil, nil, err
		}

		var stmtInstructions []*execution.InstructionExecution
		var stmtLoaderInstructions []*execution.InstructionExecution

		switch stmtType := parsedStmt.(type) {
		case *actparser.DMLStmt:
			stmtInstructions, stmtLoaderInstructions, err = convertDml(stmtType, mut)
		case *actparser.ExtensionCallStmt:
			stmtInstructions, stmtLoaderInstructions, err = convertExtensionExecute(stmtType)
		case *actparser.ActionCallStmt:
			stmtInstructions, stmtLoaderInstructions, err = convertProcedureCall(stmtType)
		}
		if err != nil {
			return nil, nil, err
		}

		loaderInstructions = append(loaderInstructions, stmtLoaderInstructions...)
		procedureInstructions = append(procedureInstructions, stmtInstructions...)
	}

	scoping := execution.ProcedureScopingPrivate
	if procedure.Public {
		scoping = execution.ProcedureScopingPublic
	}

	return &execution.Procedure{
		Name:       procedure.Name,
		Parameters: procedure.Args,
		Scoping:    scoping,
		Body:       procedureInstructions,
	}, loaderInstructions, nil
}

func convertDml(dml *actparser.DMLStmt, mut bool) (procedureInstructions, loaderInstructions []*execution.InstructionExecution, err error) {
	uniqueName := randomHash.getRandomHash()
	loadOp := &execution.InstructionExecution{
		Instruction: execution.OpDMLPrepare,
		Args: []any{
			uniqueName,
			dml.Statement,
		},
	}

	procedureOp := &execution.InstructionExecution{
		Instruction: execution.OpDMLExecute,
		Args: []any{
			uniqueName,
		},
	}
	if !mut { // i.e. may be executed with read-only Query
		// The entire statement would be unused for mutative statements since
		// they always use a prepared statement. Executions doing a read-only
		// query for uncommittted data will not use the prepared statement.
		procedureOp.Args = append(procedureOp.Args, dml.Statement)
	}

	return []*execution.InstructionExecution{procedureOp}, []*execution.InstructionExecution{loadOp}, nil
}

func convertExtensionExecute(ext *actparser.ExtensionCallStmt) (procedureInstructions, loaderInstructions []*execution.InstructionExecution, err error) {
	var setters []*execution.InstructionExecution

	var args []string
	for _, arg := range ext.Args {
		if isIdent(arg) {
			args = append(args, arg)
			continue
		}

		loadOp, uniqueName := newSetter(arg)

		args = append(args, uniqueName)
		setters = append(setters, loadOp)
	}

	procedureOp := &execution.InstructionExecution{
		Instruction: execution.OpExtensionExecute,
		Args: []any{
			ext.Extension,
			ext.Method,
			args,
			ext.Receivers,
		},
	}

	return append(setters, procedureOp), []*execution.InstructionExecution{}, nil
}

func convertProcedureCall(action *actparser.ActionCallStmt) (procedureInstructions, loaderInstructions []*execution.InstructionExecution, err error) {
	var setters []*execution.InstructionExecution

	var args []string
	for _, arg := range action.Args {
		if isIdent(arg) {
			args = append(args, arg)
			continue
		}

		loadOp, uniqueName := newSetter(arg)

		args = append(args, uniqueName)
		setters = append(setters, loadOp)
	}

	procedureOp := &execution.InstructionExecution{
		Instruction: execution.OpProcedureExecute,
		Args: []any{
			action.Method,
			args,
		},
	}

	return append(setters, procedureOp), []*execution.InstructionExecution{}, nil
}

func isIdent(val string) bool {
	if len(val) < 2 {
		return false
	}

	if val[0] != '$' && val[0] != '@' && val[0] != '!' {
		return false
	}

	return true
}

// newSetter creates a new setter instruction and returns the name of the variable
// this is useful for anonymous variables that need to be set
func newSetter(value string) (setter *execution.InstructionExecution, varName string) {
	varName = randomHash.getRandomIdent()
	setter = &execution.InstructionExecution{
		Instruction: execution.OpSetVariable,
		Args: []any{
			varName,
			trimQuotes(value),
		},
	}

	return setter, varName
}

// trimQuotes removes the first and last single or double quotes from a string
// if the string is not quoted, it is returned as is
func trimQuotes(str string) string {
	if len(str) < 2 {
		return str
	}

	if str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}

	if str[0] == '\'' && str[len(str)-1] == '\'' {
		return str[1 : len(str)-1]
	}

	return str
}
