//go:build actions_adhoc || ext_test

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/sql"
)

/*
	This file contains an extension that allows Kwil users to execute ad-hoc SQL statements.
	This works both in regular and view actions, however view actions will not be able to
	modify data.  It is expected that query strings are passed as arguments in the action.

	It has two methods: Execute and Query.
	They both do the exact same thing, except that
		- Execute can read uncommitted data, while Query cannot
		- Execute can modify data, while Query cannot

	While it is mostly meant to be an example, it likely has some practical use cases.
	Some examples include:
		- 	an oracle implementation might want flexibility to be able to execute
			ad-hoc queries. the query string could be passed as an argument in the
			vote extension payload.
		- 	a user might want to give users ad-hoc read access based on some access
			control / authentication mechanism.
*/

const adhocName = "adhoc"

func init() {
	a := &adhocExtension{}
	err := RegisterLegacyExtension(adhocName, a)
	if err != nil {
		panic(err)
	}
}

// adhocExtension is an extension that is not registered with the extension registry.
// It allows execution of ad-hoc SQL statements in the engine.
// It will return results to the engine.
type adhocExtension struct{}

// Has two methods: Execute and Query.
// We check that only one argument is passed, and that it is a string.
// We then execute the query against the datastore.
func (a *adhocExtension) Execute(scope *execution.ProcedureContext, metadata map[string]string, method string, args ...any) ([]any, error) {
	lowerMethod := strings.ToLower(method)
	if len(args) != 1 {
		return nil, fmt.Errorf("adhoc: expected 1 string argument, got %d", len(args))
	}
	stmt, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("adhoc: expected string argument, got %T", args[0])
	}

	dataset, err := scope.Dataset(scope.DBID)
	if err != nil {
		return nil, err
	}

	// for either execution, we will pass the scope.Values() as the arguments.
	// this makes it possible to use @caller, etc in the ad-hoc statement.
	var res *sql.ResultSet
	switch lowerMethod {
	default:
		return nil, fmt.Errorf(`unknown method "%s"`, method)
	case "execute":
		res, err = dataset.Execute(scope.Ctx, stmt, scope.Values())
	case "query":
		res, err = dataset.Query(scope.Ctx, stmt, scope.Values())
	}
	if err != nil {
		return nil, err
	}

	// We set the result, so that if an ad-hoc read is executed in a view action,
	// the result will be returned to the engine.
	scope.Result = res

	return nil, nil
}

// Takes no initialization parameters.
func (a *adhocExtension) Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error) {
	return nil, nil
}
