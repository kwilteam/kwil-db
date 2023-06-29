package eng

import "fmt"

type Instruction string

const (
	// Setters
	OpSetVariable Instruction = "set_var" // p1: variable name. p2: variable value. sets a variable in the context.

	// DML
	OpDMLPrepare Instruction = "dml_prep" // p1: statement name. p2: statement. adds it to the prepared statement cache
	OpDMLExecute Instruction = "dml_exec" // p1: prepared_stmt name. takes a prepared statement name and executes the statement, using any variables already in the context.

	// Extension
	OpExtensionInitialize Instruction = "ext_init" // p1: uninitialized extension name. p2: name for the newly initialized extension. p3: map of config_variable name to the set variable name.  It adds the extension to the extension cache.
	OpExtensionExecute    Instruction = "ext_exec" // p1: initialized extension name. p2: method name. p3: list of argument variable names. p4: list of variable names to assign   executes the method on the extension.

	// Procedure
	OpProcedureExecute Instruction = "proc_exec" // p1: procedure name. p2: list of set variable names to pass to the procedure. executes the procedure.
)

func (o Instruction) evaluate(ctx *procedureContext, eng *Engine, args ...any) error {
	switch o {
	default:
		return fmt.Errorf("unknown instruction '%s'", o)
	case OpSetVariable:
		return evalSetVariable(ctx, eng, args...)
	case OpDMLPrepare:
		return evalDMLPrepare(ctx, eng, args...)
	case OpDMLExecute:
		return evalDMLExecute(ctx, eng, args...)
	case OpExtensionInitialize:
		return evalExtensionInitialize(ctx, eng, args...)
	case OpExtensionExecute:
		return evalExtensionExecute(ctx, eng, args...)
	case OpProcedureExecute:
		return evalProcedureExecute(ctx, eng, args...)
	}
}

func evalSetVariable(ctx *procedureContext, eng *Engine, args ...any) error {
	if len(args) != 2 {
		return fmt.Errorf("%w: set variable requires 2 arguments, got %d", ErrIncorrectNumArgs, len(args))
	}

	varName, err := getIdent(args[0])
	if err != nil {
		return err
	}

	ctx.values[varName] = args[1]
	return nil
}

func evalDMLPrepare(ctx *procedureContext, eng *Engine, args ...any) error {
	if len(args) != 2 {
		return fmt.Errorf("%w: dml prepare requires 2 arguments, got %d", ErrIncorrectNumArgs, len(args))
	}

	stmtName, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("%w: expected string, got %T", ErrIncorrectInputType, args[0])
	}

	stmt, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("%w: expected string, got %T", ErrIncorrectInputType, args[1])
	}

	previousStmt, ok := eng.cache.preparedStatements[stmtName]
	if ok {
		err := previousStmt.Close()
		if err != nil {
			return fmt.Errorf("failed to close previous statement: %w", err)
		}
	}

	preparedStmt, err := eng.db.Prepare(stmt)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	eng.cache.preparedStatements[stmtName] = preparedStmt
	return nil
}

func evalDMLExecute(ctx *procedureContext, eng *Engine, args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: dml execute requires 1 argument, got %d", ErrIncorrectNumArgs, len(args))
	}

	stmtName, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("%w: expected string, got %T", ErrIncorrectInputType, args[0])
	}

	preparedStmt, ok := eng.cache.preparedStatements[stmtName]
	if !ok {
		return fmt.Errorf("%w: '%s'", ErrUnknownPreparedStatement, stmtName)
	}

	var err error
	ctx.lastDmlResult, err = preparedStmt.Execute(ctx.ctx, ctx.values)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func evalExtensionInitialize(ctx *procedureContext, eng *Engine, args ...any) error {
	if len(args) != 3 {
		return fmt.Errorf("%w: extension initialize requires 3 arguments, got %d", ErrIncorrectNumArgs, len(args))
	}

	extensionName, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("%w: expected string for extension name, got %T", ErrIncorrectInputType, args[0])
	}

	initializedName, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("%w: expected string for initialized extension name, got %T", ErrIncorrectInputType, args[1])
	}

	config, ok := args[2].(map[string]string)
	if !ok {
		return fmt.Errorf("%w: expected map[string]string for config, got %T", ErrIncorrectInputType, args[2])
	}

	extensionInitializer, ok := eng.availableExtensions[extensionName]
	if !ok {
		return fmt.Errorf("%w: '%s'", ErrUnknownExtension, extensionName)
	}

	newConfig := make(map[string]string)
	for configKey, setVarName := range config {
		concreteValue, ok := ctx.values[setVarName]
		if !ok {
			return fmt.Errorf("%w: '%s'", ErrUnknownVariable, setVarName)
		}

		newConfig[configKey], ok = concreteValue.(string)
		if !ok {
			return fmt.Errorf("%w: expected string, got %T", ErrIncorrectInputType, concreteValue)
		}
	}

	initializedExtension, err := extensionInitializer.Initialize(ctx.ctx, newConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize extension: %w", err)
	}

	eng.cache.initializedExtensions[initializedName] = initializedExtension
	return nil
}

func evalProcedureExecute(ctx *procedureContext, eng *Engine, args ...any) error {
	if len(args) != 2 {
		return fmt.Errorf("%w: procedure execute requires 2 arguments, got %d", ErrIncorrectNumArgs, len(args))
	}

	procedureName, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("%w: expected string for procedure name, got %T", ErrIncorrectInputType, args[0])
	}

	procedure, ok := eng.procedures[procedureName]
	if !ok {
		return fmt.Errorf("%w '%s'", ErrUnknownProcedure, procedureName)
	}

	procedureArgNames, ok := args[1].([]string)
	if !ok {
		return fmt.Errorf("%w: expected []string, got %T", ErrIncorrectInputType, args[1])
	}

	var procedureArgs []any
	for _, argName := range procedureArgNames {
		argValue, ok := ctx.values[argName]
		if !ok {
			return fmt.Errorf("%w: unknown variable '%s'", ErrIncorrectInputType, argName)
		}

		procedureArgs = append(procedureArgs, argValue)
	}

	err := procedure.evaluate(ctx.executionContext, eng, procedure.Body, procedureArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute procedure: %w", err)
	}

	return nil
}

func evalExtensionExecute(ctx *procedureContext, eng *Engine, args ...any) error {
	if len(args) != 4 {
		return fmt.Errorf("%w: extension execute requires 3 arguments, got %d", ErrIncorrectNumArgs, len(args))
	}

	extensionName, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("%w: expected string for extension name, got %T", ErrIncorrectInputType, args[0])
	}

	methodName, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("%w: expected string for method name, got %T", ErrIncorrectInputType, args[1])
	}

	argumentNames, ok := args[2].([]string)
	if !ok {
		return fmt.Errorf("%w: expected []string for arguments, got %T", ErrIncorrectInputType, args[2])
	}

	for _, argName := range argumentNames {
		err := isIdent(argName)
		if err != nil {
			return err
		}
	}

	returns, ok := args[3].([]string)
	if !ok {
		return fmt.Errorf("%w: expected []string for returns, got %T", ErrIncorrectInputType, args[3])
	}

	for _, returnName := range returns {
		err := isIdent(returnName)
		if err != nil {
			return err
		}
	}

	initializedExtension, ok := eng.cache.initializedExtensions[extensionName]
	if !ok {
		return fmt.Errorf("%w: '%s'", ErrUnitializedExtension, extensionName)
	}

	var extensionArgs []any
	for _, argName := range argumentNames {
		argValue, ok := ctx.values[argName]
		if !ok {
			return fmt.Errorf("%d: '%s'", ErrUnknownVariable, argName)
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
