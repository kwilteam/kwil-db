//go:build pglive

package pg

import (
	"context"
	"testing"
)

// TestVersion tests the pgVersion function, and ensures that the versions of
// postgres specified by any test services meet the version requirements
func TestVersion(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPool(ctx, &cfg.PoolConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	conn, err := pool.writer.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	ver, verNum, err := pgVersion(ctx, conn.Conn())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ver)

	major, minor, ok := validateVersion(verNum, verMajorRequired, verMinorRequired)
	if !ok {
		t.Errorf("unsupported postgres version %d.%d", major, minor)
	}
}
