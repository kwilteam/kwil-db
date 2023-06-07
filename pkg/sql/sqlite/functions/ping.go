package functions

import (
	"net"
	"time"

	"github.com/kwilteam/go-sqlite"
)

// impl is the implementation of the ERROR function.
var pingImpl = &SQLiteFunc{
	NArgs:         0,
	Deterministic: true,
	AllowIndirect: true,
	Scalar:        pingGoogle,
}

func pingGoogle(ctx sqlite.Context, args []sqlite.Value) (sqlite.Value, error) {
	timeout := time.Second
	_, err := net.DialTimeout("tcp", "google.com:80", timeout)

	if err != nil {
		panic(err)
	}

	return sqlite.IntegerValue(1), nil
}
