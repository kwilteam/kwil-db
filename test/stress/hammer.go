package main

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/gatewayclient"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"go.uber.org/zap"
)

// runLooped executes a basic function with a specified delay between each call
// (note this is not a ticker, which would attempt to keep a regular interval).
// If the function has an error, it is only logged. The function should be a
// closure, getting it's inputs and assigning its outputs in the scope of the
// caller.
func runLooped(ctx context.Context, fn func() error, name string, every time.Duration, logger *log.Logger) {
	defer wg.Done()
	if every < 0 { // so caller doesn't need to put an if around this
		return
	}
	for {
		err := fn()
		if err != nil {
			logger.Warn(fmt.Sprintf("%s error: %v", name, err))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(every): // not a ticker, but a wait between runs
		}
	}
}

// hammer is a high level function to begin certain programs designed to
// simulate high utilization. This includes using a test harness type to execute
// actions and make requests with a Kwil new user (unless key is specified),
// using freshly deployed toy datasets that are embedded into this tool. We may
// want to run multiple of these in concurrent goroutines in the future.
func hammer(ctx context.Context) error {
	var err error
	var priv *crypto.Secp256k1PrivateKey
	if key == "" { // only useful with no gas or when spamming non-tx/view calls
		priv, err = crypto.GenerateSecp256k1Key()
		if err != nil {
			return err
		}
		fmt.Printf("Generated new key: %v\n\n", priv.Hex())
	} else {
		priv, err = crypto.Secp256k1PrivateKeyFromHex(key)
		if err != nil {
			return err
		}
	}
	signer := &auth.EthPersonalSigner{Key: *priv}
	acctID := signer.Identity()
	fmt.Println("Identity:", hex.EncodeToString(acctID))

	logger := log.New(log.Config{
		Level:       log.InfoLevel.String(),
		OutputPaths: []string{"stdout"},
		Format:      log.FormatPlain,
		EncodeTime:  log.TimeEncodingEpochMilli, // for readability, log.TimeEncodingRFC3339Milli
	})
	logger = *logger.WithOptions(zap.AddStacktrace(zap.FatalLevel))
	trLogger := *logger.WithOptions(zap.AddCallerSkip(1))

	var kwilClt common.Client // probably should be moved to core
	if gatewayProvider {
		kwilClt, err = gatewayclient.NewClient(ctx, host, &gatewayclient.GatewayOptions{
			ClientOptions: client.ClientOptions{
				Signer:  signer,
				ChainID: chainId,
				Logger:  trLogger,
			},
		})
	} else {
		kwilClt, err = client.NewClient(ctx, host, &client.ClientOptions{
			Signer:  signer,
			ChainID: chainId,
			Logger:  trLogger,
		})
	}

	if err != nil {
		return err
	}

	kwilClt = &timedClient{Client: kwilClt, logger: &logger, showReqDur: rpcTiming}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // any early return cancels other goroutines

	_, err = kwilClt.Ping(ctx)
	if err != nil {
		return err
	}

	// Bring up the DB test harness with a fresh test database.
	h := &harness{
		Client:              kwilClt,
		concurrentBroadcast: !sequentialBroadcast,
		logger:              &logger,
		acctID:              acctID,
		signer:              signer,
		nestedLogger:        logger.WithOptions(zap.AddCallerSkip(1)),
	}

	if acct, err := kwilClt.GetAccount(ctx, acctID, types.AccountStatusPending); err != nil {
		return err
	} else { // scoping acct
		fmt.Println(acct)
		h.nonce = acct.Nonce
	}

	dbid, err := h.deployDB(ctx) // getOrCreateDB(ctx)
	if err != nil {
		return err
	}

	// ## badgering read-only requests to various systems

	// bother the account store
	wg.Add(1)
	go runLooped(ctx, func() error {
		_, err := h.GetAccount(ctx, acctID, types.AccountStatus(rand.Intn(3)))
		return err
	}, "GetAccount", badgerInterval, &logger)

	wg.Add(1)
	go runLooped(ctx, func() error {
		notAnAccount := randomBytes(len(acctID))
		_, err := h.GetAccount(ctx, notAnAccount, types.AccountStatusPending)
		return err
	}, "GetAccount", badgerInterval, &logger)

	// bother the masterDB
	wg.Add(1)
	go runLooped(ctx, func() error {
		_, err := h.ListDatabases(ctx, h.signer.Identity())
		return err
	}, "ListDatabases", badgerInterval, &logger)

	// bother the authn&kgw
	wg.Add(1)
	go runLooped(ctx, func() error {
		_, err := kwilClt.CallAction(ctx, dbid, actAuthnOnly, []any{})
		return err
	}, "call authn action", viewInterval, &logger)

	// ## "deploy / drop" program - trivial deploy/drop cycle, sometimes
	// immediately dropping. The interval for this one is different since it is
	// a delay after the drop tx confirms before deploying the next, so an
	// interval of 0 is more sensible. Should be updated with an action or two.

	wg.Add(1)
	go runLooped(ctx, func() error {
		newDBID, promise, err := h.deployDBAsync(ctx)
		if err != nil {
			return err
		}
		h.printf("deploying temp db %v", newDBID)

		if noDrop {
			return nil
		}

		dropNow := fastDropRate > 0 && rand.Intn(fastDropRate) == 0
		if dropNow {
			// drop now
			h.printf("immediately dropping new db %s", newDBID)
			err = errors.Join(h.dropDB(ctx, newDBID))
		}

		// TODO: in the deploy/drop scenario, there is a lot more to exercise
		// with actions (both view and mutable) here.

		res := <-promise
		if resErr := res.Error(); resErr != nil {
			return errors.Join(err, resErr) // deploy failed, no drop needed
		}
		h.printf("deployed temp db %v", newDBID)

		if dropNow {
			return nil // already dropped it
		}

		h.printf("dropping temp db %s", newDBID)
		return h.dropDB(ctx, newDBID)
	}, "deploy/drop", deployDropInterval, &logger)

	// ## "posters" exec/view program - work with the toy social media scheme,
	// concurrently posting and retrieving random posts.

	var pid atomic.Int64 // post ID accessed by separate goroutines

	userID, userName, err := h.getOrCreateUser(ctx, dbid)
	if err != nil {
		return fmt.Errorf("getOrCreateUser: %w", err)
	}
	h.printf("user ID = %d / user name = %v", userID, userName)

	nextPostId, err := h.nextPostId(ctx, dbid, userID)
	if err != nil {
		return fmt.Errorf("nextPostId: %w", err)
	}
	h.printf("next post ID = %d", nextPostId)
	pid.Store(int64(nextPostId))

	wg.Add(1)
	go runLooped(ctx, func() error {
		postID := strconv.Itoa(rand.Intn(int(pid.Load() + 1)))
		_, err := kwilClt.CallAction(ctx, dbid, actGetPost, []any{postID})
		if err != nil {
			return err
		}
		return nil
	}, "get post", viewInterval, &logger)

	// post

	if maxPosters > 0 {
		// Content length is limited by multiple things: message size, max transaction size, block size e.g.:
		//  - "rpc error: code = ResourceExhausted desc = grpc: received message larger than max (5000168 vs. 4194304)"
		//  - "Tx too large. Max size is 1048576, but got 4192304" a little less than 1MiB would be 1<<20 - 1e3
		bigData := randomBytes(maxContentLen) // pregenerate some random data for post content

		posters := make(chan struct{}, maxPosters)
		wg.Add(1)
		go runLooped(ctx, func() error {
			posters <- struct{}{}
			t0 := time.Now()
			next := int(pid.Add(1))
			defer func() {
				since := time.Since(t0)
				var slow string
				if since > 200*time.Millisecond {
					slow = " (SLOW)"
				}
				h.printf("new post id = %d, took %vms%s", next, float64(since.Microseconds())/1e3, slow)
			}()

			content := string(bigData[:rand.Intn(maxContentLen)+1]) // random.String(rand.Intn(maxContentLen) + 1) // randomBytes(maxContentLen)
			h.printf("beginning createPostAsync id = %d, content len = %d (concurrent with %d others)",
				next, len(content), len(posters)-1)
			promise, err := h.createPostAsync(ctx, dbid, next, "title_"+strconv.Itoa(next), content)
			if err != nil {
				<-posters
				return err
			}

			go func() {
				timer := time.NewTimer(10 * time.Second)
				defer timer.Stop()
				select {
				case res := <-promise:
					if err := res.Error(); err != nil {
						h.printf("createPost failed: %v", err)
					}
				case <-timer.C:
					logger.Error("timed out waiting for create post tx to be mined")
				}

				<-posters
			}()
			return nil
		}, "create post", postInterval, &logger)
	}

	// Some other things to consider

	// action delete_post($id) public
	// action delete_user_by_id ($id) public owner
	// action delete_user() public
	// action update_user($id, $username, $age) public

	// action list_users() public
	// action get_user_posts($username) public
	// action get_user_posts_by_userid($id) public
	// action multi_select() public

	wg.Wait()

	return nil
}

func randomBytes(l int) []byte {
	b := make([]byte, l)
	_, _ = crand.Read(b)
	return b
}
