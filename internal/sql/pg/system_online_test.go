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
	conn := pool.writer

	ver, verNum, err := pgVersion(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ver)

	major, minor, ok := validateVersion(verNum, verMajorRequired, verMinorRequired)
	if !ok {
		t.Errorf("unsupported postgres version %d.%d", major, minor)
	}
}
