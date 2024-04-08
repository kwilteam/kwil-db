package main

// This embeds the toy schema required by the test harness.

import (
	_ "embed"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/kuneiform"
)

//go:embed scheme.kf
var testScheme string

// These are the actions currently used by the harness.
const (
	actGetPost      = "get_post"
	actCreatePost   = "create_post"
	actCreateUser   = "create_user"
	actListUsers    = "list_users"
	actGetUserPosts = "get_user_posts_by_userid"
	actAuthnOnly    = "authn_only" // matters with KGW
)

func loadTestSchema() (*types.Schema, error) {
	return kuneiform.Parse(testScheme)
}

func init() {
	// Consistency check the embedded schema
	schema, err := loadTestSchema()
	if err != nil {
		panic(fmt.Sprintf("bad test schema: %v", err))
	}
	haveActions := make(map[string]bool, len(schema.Actions))
	for _, act := range schema.Actions {
		haveActions[act.Name] = true
	}
	for _, expected := range []string{actGetPost, actCreatePost, actCreateUser,
		actListUsers, actGetUserPosts, actAuthnOnly} {
		if !haveActions[expected] {
			panic(fmt.Sprintf("missing action %v", expected))
		}
	}
}
