package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/kwilteam/kwil-db/core/utils/random"
)

// The harness methods in this file pertain to the embedded dataset schema,
// testScheme.

type asyncResp struct {
	err error
	res *transactions.TransactionResult
}

func (ar *asyncResp) Error() error {
	if ar.err != nil {
		return ar.err
	}
	if ar.res.Code != 0 {
		return fmt.Errorf("execution failed with code %d, log: %q",
			ar.res.Code, ar.res.Log)
	}
	return nil
}

func (h *harness) dropDB(ctx context.Context, dbid string) error {
	var txHash transactions.TxHash
	err := h.underNonceLock(ctx, func(nonce int64) error {
		var err error
		txHash, err = h.DropDatabaseID(ctx, dbid, client.WithNonce(nonce))
		return err
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	txResp, err := h.WaitTx(ctx, txHash, txPollInterval)
	if err != nil {
		err = errors.Join(err, h.recoverNonce(ctx))
		return fmt.Errorf("WaitTx (drop): %w", err)
	}
	if code := txResp.TxResult.Code; code != 0 {
		return fmt.Errorf("drop tx failed (%d): %v", code, txResp.TxResult.Log)
	}
	return nil
}

func (h *harness) deployDBAsync(ctx context.Context) (string, <-chan asyncResp, error) {
	schema, err := loadTestSchema()
	if err != nil {
		return "", nil, err
	}
	schema.Name = random.String(12)

	var txHash transactions.TxHash
	err = h.underNonceLock(ctx, func(nonce int64) error {
		var err error
		txHash, err = h.DeployDatabase(ctx, schema, client.WithNonce(nonce))
		return err
	})
	if err != nil {
		return "", nil, err
	}

	dbid := utils.GenerateDBID(schema.Name, h.Signer.Identity())

	promise := make(chan asyncResp, 1)
	go func() {
		// time.Sleep(500 * time.Millisecond) // lame, see executeAction notes
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		resp, err := h.WaitTx(ctx, txHash, txPollInterval)
		if err != nil {
			err = errors.Join(err, h.recoverNonce(ctx))
			promise <- asyncResp{err: err}
			return
		}
		promise <- asyncResp{res: &resp.TxResult}
		h.logger.Info(fmt.Sprintf("database %q deployed in block %d", dbid, resp.Height))
	}()

	return dbid, promise, nil
}

func (h *harness) deployDB(ctx context.Context) (string, error) {
	dbid, promise, err := h.deployDBAsync(ctx)
	if err != nil {
		return "", err
	}
	res := <-promise
	if res.err != nil {
		return "", err
	}
	txRes := res.res
	if code := txRes.Code; code != 0 {
		return "", fmt.Errorf("failed to deploy database, code = %d, log = %q", code, txRes.Log)
	}
	return dbid, nil
}

func (h *harness) getOrCreateUser(ctx context.Context, dbid string) (int, string, error) {
	const (
		meUser     = "me"
		molAge int = 42
	)

	recs, err := h.CallAction(ctx, dbid, actListUsers, nil)
	if err != nil {
		return 0, "", fmt.Errorf("%s: %w", actListUsers, err)
	}
	h.printRecs(ctx, recs)
	recs.Reset()

	var userID int
	var userName string
	for recs.Next() {
		rec := *recs.Record()
		uid, user, wallet := rec["id"].(int), rec["username"].(string), rec["wallet"].([]byte)
		if bytes.Equal(wallet, h.acctID) {
			userName = user
			userID = uid
			break
		}
	}
	if userName == "" {
		userName, userID = meUser, int(random.New().Int63())
		err := h.executeAction(ctx, dbid, actCreateUser, [][]any{{userID, meUser, molAge}})
		if err != nil {
			return 0, "", fmt.Errorf("%s: %w", actCreateUser, err)
		}
		h.logger.Info(fmt.Sprintf("Added me to users table: %d / %v", userID, userName))
	} else {
		h.logger.Info(fmt.Sprintf("Found me in list_users: %d / %v", userID, userName))
	}

	return userID, userName, nil
}

func (h *harness) nextPostId(ctx context.Context, dbid string, userID int) (int, error) {
	recs, err := h.CallAction(ctx, dbid, actGetUserPosts, []any{userID}) // tuples for the Schema.Actions[i].Inputs
	if err != nil {
		return 0, fmt.Errorf("get_user_posts_by_userid: %w", err)
	}
	h.printRecs(ctx, recs)
	var nextPostId int
	for recs.Next() {
		rec := *recs.Record()
		if postID := rec["id"].(int); postID >= nextPostId {
			nextPostId = postID + 1
		}
	}
	return nextPostId, nil
}

// createPost is the synchronous version of createPostAsync.  It's unused
// presently, but this whole thing is a playground, so it remains for now.
/* xxx
func (h *harness) createPost(ctx context.Context, dbid string, postID int, title, content string) error {
	err := h.executeAction(ctx, dbid, actCreatePost, [][]any{{postID, title, content}})
	if err == nil {
		h.printf("Created post: %d / %v (len %d)", postID, title, len(content))
	}
	return err
}
*/

func (h *harness) createPostAsync(ctx context.Context, dbid string, postID int, title, content string) (<-chan asyncResp, error) {
	txHash, err := h.executeActionAsync(ctx, dbid, actCreatePost, [][]any{{postID, title, content}})
	if err != nil {
		return nil, err
	}

	promise := make(chan asyncResp, 1)
	go func() {
		// time.Sleep(500 * time.Millisecond) // lame, see executeAction notes
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
		promise <- asyncResp{res: &resp.TxResult}
	}()
	return promise, nil
}
