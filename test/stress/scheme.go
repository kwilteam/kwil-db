package main

// This embeds the toy schema required by the test harness.

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/random"
	"github.com/kwilteam/kwil-db/parse"
)

//go:embed actions.kf
var testSchemaContents string

var testSchema *types.Schema

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
	return parse.Parse([]byte(testSchemaContents))
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
	testSchema = schema
}

type actSchemaClient struct {
	h *harness
}

func (asc *actSchemaClient) deployDB(ctx context.Context) (string, error) {
	return asc.h.deployDB(ctx, testSchema)
}

func (asc *actSchemaClient) deployDBAsync(ctx context.Context) (string, <-chan asyncResp, error) {
	return asc.h.deployDBAsync(ctx, testSchema)
}

func (asc *actSchemaClient) getOrCreateUser(ctx context.Context, dbid string) (int, string, error) {
	const (
		meUser     = "me"
		molAge int = 42
	)
	// fmt.Println("dbid", dbid)

	h := asc.h

	recs, err := h.Call(ctx, dbid, actListUsers, nil)
	if err != nil {
		return 0, "", fmt.Errorf("%s: %w", actListUsers, err)
	}
	h.printRecs(ctx, recs)
	recs.Reset()

	var userID int
	var userName string
	for recs.Next() {
		rec := recs.Record()
		uid, user, wallet := rec["id"].(int64), rec["username"].(string), rec["wallet"].(string)
		acctID, err := hex.DecodeString(strings.TrimPrefix(wallet, "0x"))
		if err != nil {
			return 0, "", fmt.Errorf("bad wallet account %q: %w", wallet, err)
		}
		if bytes.Equal(acctID, h.acctID) {
			userName = user
			userID = int(uid)
			break
		}
	}
	if userName == "" {
		userName, userID = meUser, int(random.New().Int63())
		err := h.execute(ctx, dbid, actCreateUser, [][]any{{userID, meUser, molAge}})
		if err != nil {
			return 0, "", fmt.Errorf("%s: %w", actCreateUser, err)
		}
		h.logger.Info(fmt.Sprintf("Added me to users table: %d / %v", userID, userName))
	} else {
		h.logger.Info(fmt.Sprintf("Found me in list_users: %d / %v", userID, userName))
	}

	return userID, userName, nil
}

func (asc *actSchemaClient) nextPostID(ctx context.Context, dbid string, userID int) (int, error) {
	h := asc.h
	recs, err := h.Call(ctx, dbid, actGetUserPosts, []any{userID}) // tuples for the Schema.Actions[i].Inputs
	if err != nil {
		return 0, fmt.Errorf("get_user_posts_by_userid: %w", err)
	}
	h.printRecs(ctx, recs)
	var nextPostID int
	for recs.Next() {
		rec := recs.Record()
		if postID := rec["id"].(int); postID >= nextPostID {
			nextPostID = postID + 1
		}
	}
	return nextPostID, nil
}

var ErrExpected = errors.New("expected")

func (asc *actSchemaClient) createPostAsync(ctx context.Context, dbid string, postID int, title, content string) (<-chan asyncResp, error) {
	h := asc.h
	args := [][]any{{postID, title, content}}
	// Randomly fail execution. TODO: make frequency flag, like execFailRate,
	// but we really don't need it to succeed, only be mined. The failures
	// ensure expected nonce and balance updates regardless.
	expectFail := !noErrActs && rand.Intn(6) == 0
	if expectFail {
		if rand.Intn(2) == 0 {
			// kwild.abci: "msg":"failed to execute transaction","error":"ERROR: invalid input syntax for type bigint: \"not integer\" (SQLSTATE 22P02)"
			args = [][]any{{"not integer", title, content}} // id not integer (SQL exec error)
		} else {
			// kwild.abci: "msg":"failed to execute transaction","error":"incorrect number of arguments: procedure \"create_post\" requires 3 arguments, but 2 were provided"
			args = [][]any{{postID, title}} // too few args (engine procedure call error)
		}
	}
	txHash, err := h.executeAsync(ctx, dbid, actCreatePost, args)
	if err != nil {
		if expectFail {
			err = errors.Join(err, ErrExpected)
		}
		return nil, err
	}

	promise := make(chan asyncResp, 1)
	go func() {
		t0 := time.Now()
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		resp, err := h.WaitTx(ctx, txHash, txPollInterval)
		if err != nil {
			err = errors.Join(err, h.recoverNonce(ctx))
			promise <- asyncResp{err: err}
			return
		}
		h.printf("Created post: %d / %v (len %d), waited %v", postID, title, len(content), time.Since(t0))
		promise <- asyncResp{res: &resp.TxResult, expectFail: expectFail}
	}()
	return promise, nil
}
