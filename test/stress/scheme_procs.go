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
	"github.com/kwilteam/kwil-db/parse"
)

//go:embed proc_social.kf
var testProcsSchemaContent string

var testProcsSchema *types.Schema

// These are the procedures that are currently used by the harness.
const (
	procGetPosts     = "get_recent_posts_by_size"
	procCreatePost   = "create_post"
	procCreateUser   = "create_user"
	procGetUser      = "get_user"
	procGetUserPosts = "get_recent_posts"

	procActListUsers = "list_users"
)

func loadTestProcSchema() (*types.Schema, error) {
	return parse.Parse([]byte(testProcsSchemaContent))
}

func init() {
	// Consistency check the embedded schema
	schema, err := loadTestProcSchema()
	if err != nil {
		panic(fmt.Sprintf("bad test schema: %v", err))
	}
	haveProcs := make(map[string]bool, len(schema.Procedures))
	for _, proc := range schema.Procedures {
		haveProcs[proc.Name] = true
	}
	for _, expected := range []string{procGetPosts, procCreatePost, procCreateUser,
		procGetUser, procGetUserPosts} {
		if !haveProcs[expected] {
			panic(fmt.Sprintf("missing procedure %v", expected))
		}
	}
	haveActions := make(map[string]bool, len(schema.Actions))
	for _, proc := range schema.Actions {
		haveActions[proc.Name] = true
	}
	for _, expected := range []string{procActListUsers} {
		if !haveActions[expected] {
			panic(fmt.Sprintf("missing action %v", expected))
		}
	}
	testProcsSchema = schema
}

type procSchemaClient struct {
	h *harness
}

func (psc *procSchemaClient) deployDB(ctx context.Context) (string, error) {
	return psc.h.deployDB(ctx, testProcsSchema)
}

// func (psc *procSchemaClient) deployDBAsync(ctx context.Context) (string, <-chan asyncResp, error) {
// 	return psc.h.deployDBAsync(ctx, testProcsSchema)
// }

func (psc *procSchemaClient) getUser(ctx context.Context, dbid string) (string, error) {
	const (
		meUser     = "me"
		molAge int = 42
	)
	// fmt.Println("dbid", dbid)

	h := psc.h

	var userID string

	// if the user does not exist, it will be the following error:
	//   "ERROR: user \"me\" not found (SQLSTATE P0001)"
	recs, err := h.Call(ctx, dbid, procGetUser, []any{meUser})
	if err != nil {
		if strings.Contains(err.Error(), "not found (SQLSTATE P0001)") {
			return "", nil // user does not exist
		}
		return "", fmt.Errorf("%s: %w", procGetUser, err)
	}

	recs.Reset()

	for recs.Next() {
		rec := recs.Record()
		uid, _ /*age*/ := rec["id"].(string), rec["age"].(int64)
		address, _ := rec["address"].(string), rec["post_count"].(int64)
		wallet, err := hex.DecodeString(strings.TrimPrefix(address, "0x"))
		if err != nil {
			return "", err
		}
		if bytes.Equal(wallet, h.acctID) {
			userID = uid
			break
		}
	}

	return userID, nil
}

func (psc *procSchemaClient) getOrCreateUser(ctx context.Context, dbid string) (string, string, error) {
	const (
		meUser     = "me"
		molAge int = 42
	)
	// fmt.Println("dbid", dbid)

	h := psc.h

	userName := meUser

	userID, err := psc.getUser(ctx, dbid)
	if err != nil {
		return "", "", err
	}
	if userID != "" {
		h.logger.Info(fmt.Sprintf("Found me in list_users: %s / %v", userID, userName))
		return userID, meUser, nil
	}

	err = h.execute(ctx, dbid, procCreateUser, [][]any{{userName, molAge}})
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", procCreateUser, err)
	}
	userID, err = psc.getUser(ctx, dbid)
	if err != nil {
		return "", "", err
	}
	h.logger.Info(fmt.Sprintf("Added me to users table: %s / %v", userID, userName))

	return userID, userName, nil
}

// func (psc *procSchemaClient) nextPostID(ctx context.Context, dbid string, userName string) (int, error) {
// 	h := psc.h
// 	recs, err := h.Call(ctx, dbid, procGetUserPosts, []any{userName}) // tuples for the Schema.Actions[i].Inputs
// 	if err != nil {
// 		return 0, fmt.Errorf("%s: %w", procGetUserPosts, err)
// 	}
// 	h.printRecs(ctx, recs)
// 	var nextPostID int
// 	for recs.Next() {
// 		rec := recs.Record()
// 		if postID := rec["id"].(int); postID >= nextPostID {
// 			nextPostID = postID + 1
// 		}
// 	}
// 	return nextPostID, nil
// }

func (psc *procSchemaClient) createPostAsync(ctx context.Context, dbid, content string) (<-chan asyncResp, error) {
	h := psc.h
	args := [][]any{{content}}
	// Randomly fail execution. TODO: make frequency flag, like execFailRate,
	// but we really don't need it to succeed, only be mined. The failures
	// ensure expected nonce and balance updates regardless.
	expectFail := !noErrActs && rand.Intn(6) == 0
	if expectFail {
		if false /* rand.Intn(2) == 0 */ { // temporarily disabled
			// This failure mode is disable because calling a procedure with the
			// wrong input types does not cause an error easily. At least when
			// the parameter is `text`, I can give it any type and either the
			// pgx driver or postgres backend itself stringifies it somehow.

			// kwild.abci: "msg":"failed to execute transaction","error":"ERROR: invalid input syntax for type bigint: \"not integer\" (SQLSTATE 22P02)"
			args = [][]any{{types.Uint256FromInt(12)}} // content not string (SQL exec error)
		} else {
			// kwild.abci: "msg":"failed to execute transaction","error":"incorrect number of arguments: procedure \"create_post\" requires 3 arguments, but 2 were provided"
			args = [][]any{{content, "bad"}} // too many args (engine procedure call error)
		}
	}
	txHash, err := h.executeAsync(ctx, dbid, procCreatePost, args)
	if err != nil {
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
		h.printf("Created post (len %d), waited %v", len(content), time.Since(t0))
		promise <- asyncResp{res: &resp.TxResult, expectFail: expectFail}
	}()
	return promise, nil
}
