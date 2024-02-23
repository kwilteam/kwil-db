//go:build actions_adhoc || ext_test

package actions

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
)

/*
	This file contains an extension that allows Kwil users to execute ad-hoc SQL statements.
	This works both in regular and view actions, however view actions will not be able to
	modify data.  It is expected that query strings are passed as arguments in the action.

	It has one method, "Execute", which takes a string as an argument.
	When executed, it will execute the query against the dataset, and return the result.
	If called during a blockchain tx, the query can modify the underlying dataset.

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
	err := RegisterExtension(adhocName, InitializeAdhoc)
	if err != nil {
		panic(err)
	}
}

// Takes no initialization parameters.
func InitializeAdhoc(ctx *DeploymentContext, service *common.Service, metadata map[string]string) (ExtensionNamespace, error) {
	return &adhocExtension{}, nil
}

// adhocExtension is an extension that is not registered with the extension registry.
// It allows execution of ad-hoc SQL statements in the engine.
// It will return results to the engine.
type adhocExtension struct{}

// Has one method: Execute
// We check that only one argument is passed, and that it is a string.
// We then execute the query against the datastore.
func (adhocExtension) Call(scope *ProcedureContext, app *common.App, method string, inputs []any) ([]any, error) {
	if len(inputs) != 1 {
		return nil, fmt.Errorf("adhoc: expected 1 string argument, got %d", len(inputs))
	}
	stmt, ok := inputs[0].(string)
	if !ok {
		return nil, fmt.Errorf("adhoc: expected string argument, got %T", inputs[0])
	}

	// we will pass the scope.Values() as the arguments.
	// this makes it possible to use @caller, etc in the ad-hoc statement.
	if strings.ToLower(method) != "execute" {
		return nil, fmt.Errorf(`adhoc: unknown method "%s"`, method)
	}

	res, err := app.Engine.Execute(scope.Ctx, app.DB, scope.DBID, stmt, scope.Values())
	if err != nil {
		return nil, err
	}

	// We set the result, so that if an ad-hoc read is executed in a view action,
	// the result will be returned to the engine.
	scope.Result = res

	return nil, nil
}
