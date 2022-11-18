package lease

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"kwil/x/cfgx"
	"os"
	"strings"
	"sync"
	"testing"
)

func Test_Lease(t *testing.T) {
	if t == nil {
		t.Log("## Skipping test: Test_Lease ##")
		return // intentionally ignore this test for normal ops
	}

	err := os.Setenv(cfgx.Root_Dir_Env, "./")
	if err != nil {
		t.Fatal(err)
	}

	cfg := cfgx.GetConfig().Select("db-settings")

	host := cfg.String("host")
	port := cfg.Int32("port", 5432)
	user := cfg.String("user")
	password := cfg.String("password")
	database := cfg.String("database")
	driver := cfg.GetString("driver", "postgres")

	var ssl string
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		ssl = "disable"
	} else {
		ssl = "require"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s", host, port, user, password, database, ssl)

	db, err := sql.Open(driver, psqlInfo)
	if err != nil {
		t.Fatal(err)
	}

	a, err := NewAgent(db, "lease_test")
	if err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	t.Run("Test_Lease", func(t *testing.T) {
		err = a.Subscribe(context.Background(), "test", Subscriber{
			OnAcquired: func(l Lease) {
				t.Log("acquired lease")
				wg.Done()
			},
			OnFatalError: func(err error) {
				fmt.Println("fatal error2: ", err)
				wg.Done()
				t.Fatal(err)
			},
		})

		if err != nil {
			t.Fatal(err)
		}
	})

	//t.Run("test1", func(t *testing.T) {
	//	err = a.Subscribe(context.Background(), "test", Subscriber{
	//		OnAcquired: func(l Lease) {
	//			t.Log("did not acquired lease")
	//			wg.Done()
	//		},
	//		OnFatalError: func(err error) {
	//			fmt.Println("fatal error2: ", err)
	//			wg.Done()
	//			t.Fatal(err)
	//		},
	//	})
	//
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//})

	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()
}
