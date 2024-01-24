package main

// This embeds the toy schema required by the test harness.

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kuneiform/kfparser"
	"github.com/kwilteam/kwil-db/core/types/transactions"
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

func loadTestSchema() (*transactions.Schema, error) {
	astSchema, err := kfparser.Parse(testScheme)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	schemaJson, err := astSchema.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var db transactions.Schema
	err = json.Unmarshal(schemaJson, &db)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema json: %w", err)
	}

	return &db, nil
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
