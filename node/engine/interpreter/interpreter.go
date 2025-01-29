package interpreter

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"regexp"
	"strings"
	"sync"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/validation"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/extensions/hooks"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/engine"
	"github.com/kwilteam/kwil-db/node/engine/parse"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// ThreadSafeInterpreter is a thread-safe interpreter.
// It is defined as a separate struct because there are time where
// the interpreter recursively calls itself, and we need to avoid
// deadlocks.
type ThreadSafeInterpreter struct {
	mu sync.RWMutex
	i  *baseInterpreter
}

// lock locks the interpreter with either a read or write lock, depending on the access mode of the database.
func (t *ThreadSafeInterpreter) lock(db sql.DB) (unlock func(), err error) {
	am, ok := db.(sql.AccessModer)
	if !ok {
		return nil, fmt.Errorf("database does not implement AccessModer")
	}

	if am.AccessMode() == sql.ReadOnly {
		t.mu.RLock()
		return t.mu.RUnlock, nil
	}

	t.mu.Lock()
	return t.mu.Unlock, nil
}

func (t *ThreadSafeInterpreter) Call(ctx *common.EngineContext, db sql.DB, namespace string, action string, args []any, resultFn func(*common.Row) error) (*common.CallResult, error) {
	unlock, err := t.lock(db)
	if err != nil {
		return nil, err
	}
	defer unlock()

	return t.i.call(ctx, db, namespace, action, args, resultFn, true)
}

func (t *ThreadSafeInterpreter) Execute(ctx *common.EngineContext, db sql.DB, statement string, params map[string]any, fn func(*common.Row) error) error {
	unlock, err := t.lock(db)
	if err != nil {
		return err
	}
	defer unlock()

	return t.i.execute(ctx, db, statement, params, fn, true)
}

func (t *ThreadSafeInterpreter) ExecuteWithoutEngineCtx(ctx context.Context, db sql.DB, statement string, params map[string]any, fn func(*common.Row) error) error {
	return t.i.execute(newInvalidEngineCtx(ctx), db, statement, params, fn, false)
}

// recursiveInterpreter is an interpreter that can call itself.
// It is used for extensions that need to call back into the interpreter.
type recursiveInterpreter struct {
	i *baseInterpreter
	// logs is the slice of logs that the interpreter has written.
	// It references the slice that will be returned to the caller.
	logs *[]string
}

func (r *recursiveInterpreter) Call(ctx *common.EngineContext, db sql.DB, namespace string, action string, args []any, resultFn func(*common.Row) error) (*common.CallResult, error) {
	res, err := r.i.call(ctx, db, namespace, action, args, resultFn, false)
	if err != nil {
		return nil, err
	}

	*r.logs = append(*r.logs, res.Logs...)
	return res, nil
}

func (r *recursiveInterpreter) Execute(ctx *common.EngineContext, db sql.DB, statement string, params map[string]any, fn func(*common.Row) error) error {
	return r.i.execute(ctx, db, statement, params, fn, false)
}

func (r *recursiveInterpreter) ExecuteWithoutEngineCtx(ctx context.Context, db sql.DB, statement string, params map[string]any, fn func(*common.Row) error) error {
	return r.i.execute(newInvalidEngineCtx(ctx), db, statement, params, fn, false)
}

// newInvalidEngineCtx creates a new engine context that is invalid.
// It is used with ExecuteWithoutEngineCtx to allow extensions to interact with the engine
// without being in a transaction.
func newInvalidEngineCtx(ctx context.Context) *common.EngineContext {
	return &common.EngineContext{
		TxContext: &common.TxContext{
			Ctx: ctx,
			BlockContext: &common.BlockContext{
				ChainContext: &common.ChainContext{
					NetworkParameters: &common.NetworkParameters{},
					MigrationParams:   &common.MigrationContext{},
				},
			},
		},
		OverrideAuthz: true,
		InvalidTxCtx:  true,
	}
}

// a namespace is a collection of tables and actions.
// It is conceptually equivalent to Postgres's schema, but is given a
// different name to avoid confusion.
type namespace struct {
	// availableFunctions is a map of both built-in functions and user-defined PL/pgSQL functions.
	// When the interpreter planner is created, it will be populated with all built-in functions,
	// and then it will be updated with user-defined functions, effectively allowing users to override
	// some function name with their own implementation. This allows Kwil to add new built-in
	// functions without worrying about breaking user schemas.
	// This will not include aggregate and window functions, as those can only be used in SQL.
	// availableFunctions maps local action names to their execution func.
	availableFunctions map[string]*executable
	tables             map[string]*engine.Table

	// onDeploy is called exactly once when the namespace is deployed.
	// It is used to set up the namespace.
	onDeploy func(ctx *executionContext) error
	// onUndeploy is called exactly once when the namespace is undeployed.
	// It is used to clean up the namespace.
	onUndeploy func(ctx *executionContext) error

	// namespaceType is the type of namespace.
	// It can be user-created, built-in, or extension.
	namespaceType namespaceType
	// methods is a map of methods that are available if the namespace is an extension.
	methods map[string]precompileExecutable
}

// copy creates a deep copy of the namespace.
func (n *namespace) copy() *namespace {
	n2 := &namespace{
		availableFunctions: maps.Clone(n.availableFunctions),
		tables:             make(map[string]*engine.Table), // we need to copy the tables as well, so shallow copy is not enough
		onDeploy:           n.onDeploy,
		onUndeploy:         n.onUndeploy,
		namespaceType:      n.namespaceType,
		methods:            make(map[string]precompileExecutable), // we need to copy the methods as well, so shallow copy is not enough
	}

	for tblName, tbl := range n.tables {
		n2.tables[tblName] = tbl.Copy()
	}

	for k, v := range n.methods {
		n2.methods[k] = *v.copy()
	}

	return n2
}

type namespaceType string

const (
	namespaceTypeUser      namespaceType = "USER"
	namespaceTypeSystem    namespaceType = "SYSTEM"
	namespaceTypeExtension namespaceType = "EXTENSION"
)

func (n namespaceType) valid() bool {
	switch n {
	case namespaceTypeUser, namespaceTypeSystem, namespaceTypeExtension:
		return true
	default:
		return false
	}
}

type nilNamespaceRegister struct{}

func (n nilNamespaceRegister) RegisterNamespace(ns string) {}

func (n nilNamespaceRegister) UnregisterAllNamespaces() {}

func (n nilNamespaceRegister) Lock() {}

func (n nilNamespaceRegister) Unlock() {}

// NewInterpreter creates a new interpreter.
// It reads currently stored namespaces and loads them into memory.
func NewInterpreter(ctx context.Context, db sql.DB, service *common.Service, accounts common.Accounts, validators common.Validators, nsr engine.NamespaceRegister) (*ThreadSafeInterpreter, error) {
	if nsr == nil {
		nsr = nilNamespaceRegister{}
	}

	var exists bool
	count := 0
	// we need to check if it is initialized. We will do this by checking if the schema kwild_engine exists
	err := queryRowFunc(ctx, db, "SELECT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'kwild_engine')", []any{&exists}, func() error {
		count++
		return nil
	})
	if err != nil {
		return nil, err
	}

	switch count {
	case 0:
		return nil, fmt.Errorf("could not determine if the database is initialized")
	case 1:
		if !exists {
			err = initSQL(ctx, db)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unexpected number of rows returned")
	}

	namespaces, err := listNamespaces(ctx, db)
	if err != nil {
		return nil, err
	}

	interpreter := &baseInterpreter{
		namespaces:        make(map[string]*namespace),
		service:           service,
		validators:        validators,
		accounts:          accounts,
		namespaceRegister: nsr,
	}
	interpreter.accessController, err = newAccessController(ctx, db)
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces {
		tables, err := listTablesInNamespace(ctx, db, ns.Name)
		if err != nil {
			return nil, err
		}

		tblMap := make(map[string]*engine.Table)
		for _, tbl := range tables {
			tblMap[tbl.Name] = tbl
		}

		actions, err := listActionsInBuiltInNamespace(ctx, db, ns.Name)
		if err != nil {
			return nil, err
		}

		// now, we override the built-in functions with the actions
		namespaceFunctions := copyBuiltinExecutables()
		for _, action := range actions {
			exec := makeActionToExecutable(ns.Name, action)
			namespaceFunctions[exec.Name] = exec
		}

		interpreter.namespaces[ns.Name] = &namespace{
			tables:             tblMap,
			availableFunctions: namespaceFunctions,
			namespaceType:      ns.Type,
			onDeploy:           func(ctx *executionContext) error { return nil },
			onUndeploy:         func(ctx *executionContext) error { return nil },
		}
		interpreter.accessController.registerNamespace(ns.Name)
	}

	// we need to add the tables of the info schema manually, since they are not stored in the database

	// get and initialize all used extensions
	storedExts, err := getExtensionInitializationMetadata(ctx, db)
	if err != nil {
		return nil, err
	}

	systemExtensions := precompiles.RegisteredPrecompiles()
	var instances []*precompiles.Precompile // we must call OnStart after all instances have been initialized
	for _, ext := range storedExts {
		sysExt, ok := systemExtensions[ext.ExtName]
		if !ok {
			return nil, fmt.Errorf("the database has an extension in use that is unknown to the system: %s", ext.ExtName)
		}

		namespace, inst, err := initializeExtension(ctx, service, db, sysExt, ext.Alias, ext.Metadata)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)

		// in case the extension implementation was changed, we need to update the stored method info
		err = ensureMethodsRegistered(ctx, db, ext.Alias, inst.Methods)
		if err != nil {
			return nil, err
		}

		// if a namespace already exists, we should use it instead, since it might have been read earlier, and contain
		// kuneiform actions and tables
		if existing, ok := interpreter.namespaces[ext.Alias]; ok {
			// kuneiform actions should overwrite methods,
			// so any actions already read should just overwrite the methods
			for k, v := range existing.availableFunctions {
				namespace.availableFunctions[k] = v
			}

			namespace.tables = existing.tables
		}

		interpreter.namespaces[ext.Alias] = namespace
		interpreter.accessController.registerNamespace(ext.Alias)
	}

	threadSafe := &ThreadSafeInterpreter{i: interpreter}

	app := &common.App{
		Service:    service,
		DB:         db,
		Engine:     threadSafe,
		Accounts:   accounts,
		Validators: validators,
	}

	for _, inst := range instances {
		err = inst.OnStart(ctx, app)
		if err != nil {
			return nil, err
		}
	}

	interpreter.syncNamespaceManager()

	for _, hook := range hooks.ListEngineReadyHooks() {
		err = hook(ctx, app)
		if err != nil {
			return nil, err
		}
	}

	return threadSafe, nil
}

// funcDefToExecutable converts a Postgres function definition to an executable.
// This allows built-in Postgres functions to be used within the interpreter.
// This inconveniently requires a roundtrip to the database, but it is necessary
// to ensure that the function is executed correctly. In the future, we can replicate
// the functionality of the function in Go to avoid the roundtrip. I initially tried
// to do this, however it get's extroadinarily complex when getting to string formatting.
func funcDefToExecutable(funcName string, funcDef *engine.ScalarFunctionDefinition) *executable {
	return &executable{
		Name: funcName,
		Func: func(e *executionContext, args []value, fn resultFunc) error {
			//convert args to any
			params := make([]string, len(args))
			argTypes := make([]*types.DataType, len(args))
			for i, arg := range args {
				params[i] = fmt.Sprintf("$%d", i+1)
				argTypes[i] = arg.Type()
			}

			// get the expected return type
			retTyp, err := funcDef.ValidateArgsFunc(argTypes)
			if err != nil {
				return err
			}

			// if the function name is notice, then we need to get write the notice to our logs locally,
			// instead of executing a query. This is the functional eqauivalent of Kwil's console.log().
			if funcName == "notice" {
				var log string
				if !args[0].Null() {
					log = args[0].RawValue().(string)
				}
				*e.logs = append(*e.logs, log)
				return nil
			}

			if funcName == "error" {
				var msg string
				if !args[0].Null() {
					msg = args[0].RawValue().(string)
				}
				return fmt.Errorf("error raised while executing: %s", msg)
			}

			zeroVal, err := newZeroValue(retTyp)
			if err != nil {
				return err
			}

			// format the function
			pgFormat, err := funcDef.PGFormatFunc(params)
			if err != nil {
				return err
			}

			// execute the query
			// We could avoid a roundtrip here by having a go implementation of the function.
			// Since for now we are more concerned about expanding functionality than scalability,
			// we will use the roundtrip.
			iters := 0
			err = query(e.engineCtx.TxContext.Ctx, e.db, "SELECT "+pgFormat+";", []value{zeroVal}, func() error {
				iters++
				return nil
			}, args)
			if err != nil {
				return err
			}
			if iters != 1 {
				return fmt.Errorf("expected 1 row, got %d", iters)
			}

			return fn(&row{
				columns: []string{funcName},
				Values:  []value{zeroVal},
			})
		},
		Type: executableTypeFunction,
	}
}

// baseInterpreter interprets Kwil SQL statements.
type baseInterpreter struct {
	namespaces map[string]*namespace
	// accessController is used to check if a user has access to a namespace
	accessController *accessController
	// service is the base application
	service *common.Service
	// validators is the validator manager for the application
	validators common.Validators
	// accounts is the account manager for the application
	accounts common.Accounts
	// namespaceRegister is used to register and unregister namespaces
	namespaceRegister engine.NamespaceRegister
}

// copy deep copies the state of the interpreter.
// It is used to ensure transactionality by providing
// a state that can be rolled back to.
func (i *baseInterpreter) copy() *baseInterpreter {
	namespaces := make(map[string]*namespace)
	for k, v := range i.namespaces {
		namespaces[k] = v.copy()
	}

	return &baseInterpreter{
		namespaces:       namespaces,
		accessController: i.accessController.copy(),
		// service, validators, and accounts should have no need to be copied
		service:    i.service,
		validators: i.validators,
		accounts:   i.accounts,
	}
}

// Execute executes a statement against the database.
func (i *baseInterpreter) execute(ctx *common.EngineContext, db sql.DB, statement string, params map[string]any, fn func(*common.Row) error, toplevel bool) (err error) {
	defer func() {
		i.syncNamespaceManager()
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	err = ctx.Valid()
	if err != nil {
		return err
	}

	if fn == nil {
		fn = func(*common.Row) error { return nil }
	}

	// parse the statement
	ast, err := parse.Parse(statement)
	if err != nil {
		return fmt.Errorf("%w: error in top-level statement %s: %w", engine.ErrParse, statement, err)
	}

	if len(ast) == 0 {
		return fmt.Errorf("no valid statements provided: %s", statement)
	}

	copied := i.copy()
	defer func() {
		if err != nil {
			// rollback the interpreter
			i.namespaces = copied.namespaces
			i.accessController = copied.accessController
		}
	}()

	execCtx, err := i.newExecCtx(ctx, db, engine.DefaultNamespace, toplevel)
	if err != nil {
		return err
	}

	for _, param := range order.OrderMap(params) {
		val, err := newValue(param.Value)
		if err != nil {
			return err
		}

		name := strings.ToLower(param.Key)
		if !strings.HasPrefix(name, "$") {
			name = "$" + name
		}
		if err := isValidVarName(name); err != nil {
			return err
		}

		err = execCtx.setVariable(name, val)
		if err != nil {
			return err
		}
	}

	interpPlanner := interpreterPlanner{}

	for _, stmt := range ast {
		err = stmt.Accept(&interpPlanner).(stmtFunc)(execCtx, func(row *row) error {
			return fn(rowToCommonRow(row))
		})
		if err != nil {
			return err
		}
	}

	return nil
}

var identRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// isValidVarName checks if a string is a valid variable name.
func isValidVarName(s string) error {
	if !strings.HasPrefix(s, "$") {
		return fmt.Errorf("variable name must start with $")
	}

	if !identRegexp.MatchString(s[1:]) {
		return fmt.Errorf("variable name must only contain letters, numbers, and underscores")
	}

	// we ignore the $ as part of the "name"
	if len(s[1:]) > validation.MAX_IDENT_NAME_LENGTH {
		return fmt.Errorf("variable name cannot be longer than %d characters, received %s", validation.MAX_IDENT_NAME_LENGTH, s)
	}

	return nil
}

// Call executes an action against the database.
// The resultFn is called with the result of the action, if any.
func (i *baseInterpreter) call(ctx *common.EngineContext, db sql.DB, namespace, action string, args []any, resultFn func(*common.Row) error, toplevel bool) (callRes *common.CallResult, err error) {
	defer func() {
		i.syncNamespaceManager()
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	err = ctx.Valid()
	if err != nil {
		return nil, err
	}
	if resultFn == nil {
		resultFn = func(*common.Row) error { return nil }
	}

	if namespace == "" {
		namespace = engine.DefaultNamespace
	}
	namespace = strings.ToLower(namespace)
	action = strings.ToLower(action)

	execCtx, err := i.newExecCtx(ctx, db, namespace, toplevel)
	if err != nil {
		return nil, err
	}

	ns, ok := i.namespaces[namespace]
	if !ok {
		return nil, fmt.Errorf(`namespace "%s" does not exist`, namespace)
	}

	// now we can call the executable. The executable checks that the caller is allowed to call the action
	// (e.g. in case of a private action or owner action)
	exec, ok := ns.availableFunctions[action]
	if !ok {
		return nil, fmt.Errorf(`%w: action "%s" does not exist in namespace "%s"`, engine.ErrUnknownAction, action, namespace)
	}

	switch exec.Type {
	case executableTypeFunction:
		return nil, fmt.Errorf(`action "%s" is a built-in function and cannot be called directly`, action)
	case executableTypeAction, executableTypePrecompile:
		// do nothing, this is what we want
	default:
		return nil, fmt.Errorf(`node bug: unknown executable type "%s"`, exec.Type)
	}

	argVals := make([]value, len(args))
	for i, arg := range args {
		val, err := newValue(arg)
		if err != nil {
			return nil, err
		}

		argVals[i] = val
	}

	err = exec.Func(execCtx, argVals, func(row *row) error {
		return resultFn(rowToCommonRow(row))
	})
	if err != nil {
		return nil, err
	}

	return &common.CallResult{
		Logs: *execCtx.logs,
	}, nil
}

// rowToCommonRow converts a row to a common.Row.
func rowToCommonRow(row *row) *common.Row {
	// convert the results to any
	convertedResults := make([]any, len(row.Values))
	dataTypes := make([]*types.DataType, len(row.Values))
	for i, result := range row.Values {
		convertedResults[i] = result.RawValue()
		dataTypes[i] = result.Type()
	}

	return &common.Row{
		ColumnNames: row.Columns(),
		ColumnTypes: dataTypes,
		Values:      convertedResults,
	}
}

// newExecCtx creates a new execution context.
func (i *baseInterpreter) newExecCtx(txCtx *common.EngineContext, db sql.DB, namespace string, toplevel bool) (*executionContext, error) {
	am, ok := db.(sql.AccessModer)
	if !ok {
		return nil, fmt.Errorf("database does not implement AccessModer")
	}

	logs := make([]string, 0)

	e := &executionContext{
		engineCtx:      txCtx,
		scope:          newScope(namespace),
		canMutateState: am.AccessMode() == sql.ReadWrite,
		db:             db,
		interpreter:    i,
		logs:           &logs,
	}
	e.scope.isTopLevel = toplevel

	return e, nil
}

// syncNamespaceManager syncs all current namespaces with the namespace manager.
func (i *baseInterpreter) syncNamespaceManager() {
	i.namespaceRegister.Lock()
	defer i.namespaceRegister.Unlock()
	i.namespaceRegister.UnregisterAllNamespaces()
	for ns := range i.namespaces {
		if ns == engine.InfoNamespace {
			continue
		}
		i.namespaceRegister.RegisterNamespace(ns)
	}
}

var builtInExecutables = func() map[string]*executable {
	execs := make(map[string]*executable)
	for funcName, impl := range engine.Functions {
		if scalarImpl, ok := impl.(*engine.ScalarFunctionDefinition); ok {
			execs[funcName] = funcDefToExecutable(funcName, scalarImpl)
		}
	}

	return execs
}()

// copyBuiltinExecutables returns a map of built-in functions to their executables.
func copyBuiltinExecutables() map[string]*executable {
	b := make(map[string]*executable)
	for k, v := range builtInExecutables {
		b[k] = v
	}

	return b
}
