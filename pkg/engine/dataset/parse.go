package dataset

import (
	"sync"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/engine/dataset/actparser"
	"github.com/kwilteam/kwil-db/pkg/engine/eng"
	"github.com/kwilteam/kwil-db/pkg/engine/types"

	"encoding/binary"
	"math/rand"
	"time"
)

// TODO: this file is ripe for refactoring.  Seems very bug prone

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

func getEngineOpts(procedures []*types.Procedure, extensions []*types.Extension, exts map[string]Initializer) (*eng.EngineOpts, error) {
	engineProceduresMap := make(map[string]*eng.Procedure)
	loaders := []*eng.InstructionExecution{}
	for _, procedure := range procedures {
		engineProcedure, loaderInstructions, err := parseAction(procedure)
		if err != nil {
			return nil, err
		}

		engineProceduresMap[procedure.Name] = engineProcedure
		loaders = append(loaders, loaderInstructions...)
	}

	for _, extension := range extensions {
		loaderInstructions, err := getExtensionLoader(extension)
		if err != nil {
			return nil, err
		}

		loaders = append(loaders, loaderInstructions...)
	}

	extensionMap := make(map[string]eng.Initializer)
	for name, initializer := range exts {
		extensionMap[name] = initializerWrapper{initializer}
	}

	return &eng.EngineOpts{
		Extensions: extensionMap,
		Procedures: engineProceduresMap,
		LoadCmd:    loaders,
	}, nil
}

func getExtensionLoader(extension *types.Extension) ([]*eng.InstructionExecution, error) {
	var setters []*eng.InstructionExecution
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

	initializeOp := &eng.InstructionExecution{
		Instruction: eng.OpExtensionInitialize,
		Args: []any{
			extension.Name,
			extension.Alias,
			configMap,
		},
	}

	return append(setters, initializeOp), nil
}

func parseAction(procedure *types.Procedure) (engineProcedure *eng.Procedure, loaderInstructions []*eng.InstructionExecution, err error) {
	var procedureInstructions []*eng.InstructionExecution
	loaderInstructions = []*eng.InstructionExecution{}

	for _, stmt := range procedure.Statements {
		parsedStmt, err := actparser.Parse(stmt)
		if err != nil {
			return nil, nil, err
		}

		var stmtInstructions []*eng.InstructionExecution
		var stmtLoaderInstructions []*eng.InstructionExecution

		switch stmtType := parsedStmt.(type) {
		case *actparser.DMLStmt:
			stmtInstructions, stmtLoaderInstructions, err = convertDml(stmtType)
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

	scoping := eng.ProcedureScopingPrivate
	if procedure.Public {
		scoping = eng.ProcedureScopingPublic
	}

	return &eng.Procedure{
		Name:       procedure.Name,
		Parameters: procedure.Args,
		Scoping:    scoping,
		Body:       procedureInstructions,
	}, loaderInstructions, nil
}

func convertDml(dml *actparser.DMLStmt) (procedureInstructions []*eng.InstructionExecution, loaderInstructions []*eng.InstructionExecution, err error) {
	uniqueName := randomHash.getRandomHash()

	loadOp := &eng.InstructionExecution{
		Instruction: eng.OpDMLPrepare,
		Args: []any{
			uniqueName,
			dml.Statement,
		},
	}

	procedureOp := &eng.InstructionExecution{
		Instruction: eng.OpDMLExecute,
		Args: []any{
			uniqueName,
		},
	}

	return []*eng.InstructionExecution{procedureOp}, []*eng.InstructionExecution{loadOp}, nil
}

func convertExtensionExecute(ext *actparser.ExtensionCallStmt) (procedureInstructions []*eng.InstructionExecution, loaderInstructions []*eng.InstructionExecution, err error) {
	var setters []*eng.InstructionExecution

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

	procedureOp := &eng.InstructionExecution{
		Instruction: eng.OpExtensionExecute,
		Args: []any{
			ext.Extension,
			ext.Method,
			args,
			ext.Receivers,
		},
	}

	return append(setters, procedureOp), []*eng.InstructionExecution{}, nil
}

func convertProcedureCall(action *actparser.ActionCallStmt) (procedureInstructions []*eng.InstructionExecution, loaderInstructions []*eng.InstructionExecution, err error) {
	var setters []*eng.InstructionExecution

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

	procedureOp := &eng.InstructionExecution{
		Instruction: eng.OpProcedureExecute,
		Args: []any{
			action.Method,
			args,
		},
	}

	return append(setters, procedureOp), []*eng.InstructionExecution{}, nil
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
func newSetter(value string) (setter *eng.InstructionExecution, varName string) {
	varName = randomHash.getRandomIdent()
	setter = &eng.InstructionExecution{
		Instruction: eng.OpSetVariable,
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
