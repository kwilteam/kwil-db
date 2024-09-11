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

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/gatewayclient"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"golang.org/x/sync/errgroup"

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
		EncodeTime:  log.TimeEncodingRFC3339Milli,
	})
	defer logger.Close()
	logger = *logger.WithOptions(zap.AddStacktrace(zap.FatalLevel))
	trLogger := *logger.WithOptions(zap.AddCallerSkip(1))

	var kwilClt clientType.Client
	if gatewayProvider {
		kwilClt, err = gatewayclient.NewClient(ctx, host, &gatewayclient.GatewayOptions{
			Options: clientType.Options{
				Signer:  signer,
				ChainID: chainId,
				Logger:  trLogger,
			},
		})
	} else {
		kwilClt, err = client.NewClient(ctx, host, &clientType.Options{
			Signer:            signer,
			ChainID:           chainId,
			Logger:            trLogger,
			AuthenticateCalls: authCall,
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
		concurrentBroadcast: concurrentBroadcast,
		logger:              &logger,
		acctID:              acctID,
		signer:              signer,
		nestedLogger:        logger.WithOptions(zap.AddCallerSkip(1)),
		quiet:               quiet,
	}

	if acct, err := kwilClt.GetAccount(ctx, acctID, types.AccountStatusPending); err != nil {
		return err
	} else { //nolint (scoping acct var)
		h.nonce = acct.Nonce
	}

	// procedure spammer
	psc := procSchemaClient{h}

	// action spammer
	asc := actSchemaClient{h}

	var dbid, dbidProc string
	var userIDact int
	grp, ctxg := errgroup.WithContext(ctx)
	grp.Go(func() error {
		var err error
		dbidProc, err = psc.deployDB(ctxg)
		if err != nil {
			return err
		}

		uidProc, unameProc, err := psc.getOrCreateUser(ctxg, dbidProc)
		if err != nil {
			return fmt.Errorf("getOrCreateUser: %w", err)
		}
		h.printf("proc schema: user ID = %s / user name = %v", uidProc, unameProc)
		return nil
	})
	grp.Go(func() error {
		var err error
		dbid, err = asc.deployDB(ctxg)
		if err != nil {
			return err
		}

		var userName string
		userIDact, userName, err = asc.getOrCreateUser(ctxg, dbid)
		if err != nil {
			return fmt.Errorf("getOrCreateUser: %w", err)
		}
		h.printf("act schema: user ID = %d / user name = %v", userIDact, userName)
		return nil
	})

	if err = grp.Wait(); err != nil {
		return err
	}

	h.nonceChaos = nonceChaos // after successfully deploying the test db and creating a user in it

	// ## badgering read-only requests to various systems

	// bother the account store
	wg.Add(1)
	go runLooped(ctx, func() error {
		_, err := h.GetAccount(ctx, acctID, types.AccountStatus(rand.Intn(2)))
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
		_, err := kwilClt.Call(ctx, dbid, actAuthnOnly, []any{})
		return err
	}, "call authn action", viewInterval, &logger)

	// ## "deploy / drop" program - trivial deploy/drop cycle, sometimes
	// immediately dropping. The interval for this one is different since it is
	// a delay after the drop tx confirms before deploying the next, so an
	// interval of 0 is more sensible. Should be updated with an action or two.

	wg.Add(1)
	go runLooped(ctx, func() error {
		newDBID, promise, err := asc.deployDBAsync(ctx)
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

	nextPostID, err := asc.nextPostID(ctx, dbid, userIDact)
	if err != nil {
		return fmt.Errorf("nextPostID: %w", err)
	}
	h.printf("next post ID = %d", nextPostID)
	pid.Store(int64(nextPostID))

	wg.Add(1)
	go runLooped(ctx, func() error {
		postID := strconv.Itoa(rand.Intn(int(pid.Load() + 1)))
		_, err := kwilClt.Call(ctx, dbid, actGetPost, []any{postID})
		if err != nil {
			return err
		}
		return nil
	}, "get post", viewInterval, &logger)

	// Content length is limited by multiple things: message size, max transaction size, block size e.g.:
	//  - "rpc error: code = ResourceExhausted desc = grpc: received message larger than max (5000168 vs. 4194304)"
	//  - "Tx too large. Max size is 1048576, but got 4192304" a little less than 1MiB would be 1<<20 - 1e3
	bigData := makeBigData(contentLen, h.printf)

	if maxPosters%2 != 0 {
		maxPosters++
	}

	// post (actions)
	if maxPosters > 0 {
		maxActPost := maxPosters / 2
		operation := "create post (act)"
		posters := make(chan struct{}, maxActPost)
		wg.Add(1)
		go runLooped(ctx,
			asyncFn(ctx, posters, h.printf, operation,
				func() (<-chan asyncResp, error) {
					next := int(pid.Add(1))
					var content string
					if variableLen {
						content = bigData[:rand.Intn(contentLen)+1] // random.String(rand.Intn(maxContentLen) + 1) // randomBytes(maxContentLen)
					} else {
						content = bigData[:contentLen]
					}
					h.printf("beginning createPostAsync id = %d, content len = %d (concurrent with %d others)",
						next, len(content), len(posters)-1)
					return asc.createPostAsync(ctx, dbid, next, "title_"+strconv.Itoa(next), content)
				},
			),
			operation, postInterval, &logger,
		)
	}

	// post (procedures)
	if maxPosters > 0 {
		maxProcPost := maxPosters / 2
		operation := "create post (proc)"
		posters := make(chan struct{}, maxProcPost)
		wg.Add(1)
		go runLooped(ctx,
			asyncFn(ctx, posters, h.printf, operation,
				func() (<-chan asyncResp, error) {
					var content string
					if variableLen {
						content = bigData[:rand.Intn(contentLen)+1] // random.String(rand.Intn(maxContentLen) + 1) // randomBytes(maxContentLen)
					} else {
						content = bigData[:contentLen]
					}
					h.printf("beginning createPostAsync (proc), content len = %d (concurrent with %d others)",
						len(content), len(posters)-1)
					return psc.createPostAsync(ctx, dbidProc, content)
				},
			),
			operation, postInterval, &logger,
		)
	}

	wg.Wait()

	return nil
}

func randomBytes(l int) []byte {
	b := make([]byte, l)
	_, _ = crand.Read(b)
	return b
}
