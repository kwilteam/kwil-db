package dataset2

import "fmt"

type OpCodeExecution struct {
	OpCode Instruction
	Args   []any
}

type Instruction string

const (
	// Setters
	OpCodeSetVariable Instruction = "set_var" // p1: variable name. p2: variable value. sets a variable in the context.

	// DML
	OpCodeDMLPrepare Instruction = "dml_prep" // p1: statement name. p2: statement. adds it to the prepared statement cache
	OpCodeDMLExecute Instruction = "dml_exec" // p1: prepared_stmt name. takes a prepared statement name and executes the statement, using any variables already in the context.

	// Extension
	OpCodeExtensionInitialize Instruction = "ext_init" // p1: uninitialized extension name. p2: name for the newly initialized extension. p3: map of config_variable name to the set variable name.  It adds the extension to the extension cache.
	OpCodeExtensionExecute    Instruction = "ext_exec" // p1: initialized extension name. p2: method name. p3: list of argument variable names. p4: list of variable names to assign   executes the method on the extension.

	// Procedure
	OpCodeProcedureExecute Instruction = "proc_exec" // p1: procedure name. p2: list of set variable names to pass to the procedure. executes the procedure.
)

func (o Instruction) evaluate(ctx *procedureContext, ds *Dataset, args ...any) error {
	switch o {
	default:
		return fmt.Errorf("unknown opcode '%s'", o)
	case OpCodeSetVariable:
		return evalSetVariable(ctx, ds, args...)
	case OpCodeDMLPrepare:
		return evalDMLPrepare(ctx, ds, args...)
	case OpCodeDMLExecute:
		return evalDMLExecute(ctx, ds, args...)
	case OpCodeExtensionInitialize:
		return evalExtensionInitialize(ctx, ds, args...)
	case OpCodeExtensionExecute:
		return evalExtensionExecute(ctx, ds, args...)
	case OpCodeProcedureExecute:
		return evalProcedureExecute(ctx, ds, args...)
	}
}

func evalSetVariable(ctx *procedureContext, ds *Dataset, args ...any) error {
	if len(args) != 2 {
		return fmt.Errorf("set variable requires 2 arguments, got %d", len(args))
	}

	varName, err := getIdent(args[0])
	if err != nil {
		return err
	}

	ctx.values[varName] = args[1]
	return nil
}

func evalDMLPrepare(ctx *procedureContext, ds *Dataset, args ...any) error {
	if len(args) != 2 {
		return fmt.Errorf("dml prepare requires 2 arguments, got %d", len(args))
	}

	stmtName, err := getIdent(args[0])
	if err != nil {
		return err
	}

	stmt, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", args[1])
	}

	previousStmt, ok := ds.cache.preparedStatements[stmtName]
	if ok {
		err = previousStmt.Close()
		if err != nil {
			return fmt.Errorf("failed to close previous statement: %w", err)
		}
	}

	preparedStmt, err := ds.db.Prepare(stmt)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	ds.cache.preparedStatements[stmtName] = preparedStmt
	return nil
}

func evalDMLExecute(ctx *procedureContext, ds *Dataset, args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("dml execute requires 1 argument, got %d", len(args))
	}

	stmtName, err := getIdent(args[0])
	if err != nil {
		return err
	}

	preparedStmt, ok := ds.cache.preparedStatements[stmtName]
	if !ok {
		return fmt.Errorf("unknown prepared statement '%s'", stmtName)
	}

	ctx.lastDmlResult, err = preparedStmt.Execute(ctx.values)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func evalExtensionInitialize(ctx *procedureContext, ds *Dataset, args ...any) error {
	if len(args) != 3 {
		return fmt.Errorf("extension initialize requires 3 arguments, got %d", len(args))
	}

	extensionName, err := getIdent(args[0])
	if err != nil {
		return err
	}

	initializedName, err := getIdent(args[1])
	if err != nil {
		return err
	}

	config, ok := args[2].(map[string]string)
	if !ok {
		return fmt.Errorf("expected map[string]string, got %T", args[2])
	}

	extensionInitializer, ok := ds.cache.extensionInitializers[extensionName]
	if !ok {
		return fmt.Errorf("unknown extension '%s'", extensionName)
	}

	for configKey, setVarName := range config {
		concreteValue, ok := ctx.values[setVarName]
		if !ok {
			return fmt.Errorf("unknown variable '%s'", setVarName)
		}

		config[configKey], ok = concreteValue.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", concreteValue)
		}
	}

	initializedExtension, err := extensionInitializer.Initialize(ctx.ctx, config)
	if err != nil {
		return fmt.Errorf("failed to initialize extension: %w", err)
	}

	ds.cache.initializedExtensions[initializedName] = initializedExtension
	return nil
}

func evalProcedureExecute(ctx *procedureContext, ds *Dataset, args ...any) error {
	if len(args) != 2 {
		return fmt.Errorf("procedure execute requires 2 arguments, got %d", len(args))
	}

	procedureName, err := getIdent(args[0])
	if err != nil {
		return err
	}

	procedure, ok := ds.cache.procedures[procedureName]
	if !ok {
		return fmt.Errorf("unknown procedure '%s'", procedureName)
	}

	procedureArgNames, ok := args[1].([]string)
	if !ok {
		return fmt.Errorf("expected []string, got %T", args[1])
	}

	var procedureArgs []any
	for _, argName := range procedureArgNames {
		argValue, ok := ctx.values[argName]
		if !ok {
			return fmt.Errorf("unknown variable '%s'", argName)
		}

		procedureArgs = append(procedureArgs, argValue)
	}

	err = procedure.Execute(ctx.executionContext, procedureArgs)
	if err != nil {
		return fmt.Errorf("failed to execute procedure: %w", err)
	}

	return nil
}

func evalExtensionExecute(ctx *procedureContext, ds *Dataset, args ...any) error {
	if len(args) != 4 {
		return fmt.Errorf("extension execute requires 3 arguments, got %d", len(args))
	}

	extensionName, err := getIdent(args[0]) // TODO: this should not be ident
	if err != nil {
		return err
	}

	methodName, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("expected string for method name, got %T", args[1])
	}

	argumentNames, ok := args[2].([]string)
	if !ok {
		return fmt.Errorf("expected []string for arguments, got %T", args[2])
	}

	returns, ok := args[3].([]string)
	if !ok {
		return fmt.Errorf("expected []string for returns, got %T", args[3])
	}

	for _, returnName := range returns {
		err = isIdent(returnName)
		if err != nil {
			return err
		}
	}

	initializedExtension, ok := ds.cache.initializedExtensions[extensionName]
	if !ok {
		return fmt.Errorf("unknown extension '%s'", extensionName)
	}

	var extensionArgs []any
	for _, argName := range argumentNames {
		argValue, ok := ctx.values[argName]
		if !ok {
			return fmt.Errorf("unknown variable '%s'", argName)
		}

		extensionArgs = append(extensionArgs, argValue)
	}

	results, err := initializedExtension.Execute(ctx.ctx, methodName, extensionArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute extension method: %w", err)
	}

	if len(returns) != len(results) {
		return fmt.Errorf("expected %d return values, got %d", len(returns), len(results))
	}

	for i, result := range results {
		ctx.values[returns[i]] = result
	}

	return nil
}

// utils:
func getIdent(val any) (string, error) {
	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", val)
	}

	err := isIdent(strVal)
	if err != nil {
		return "", err
	}

	return strVal, nil
}

func isIdent(val string) error {
	if len(val) < 2 {
		return fmt.Errorf("expected variable name, got '%s'", val)
	}

	if val[0] != '$' && val[0] != '@' && val[0] != '!' {
		return fmt.Errorf("expected variable name, got '%s'", val)
	}

	return nil
}
