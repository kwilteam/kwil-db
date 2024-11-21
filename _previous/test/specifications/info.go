package specifications

import (
	"context"
	"testing"
)

func ExecuteChainInfoSpecification(ctx context.Context, t *testing.T, dsl InfoDsl, chainId string) {
	t.Logf("Executing chain info specification")

	info, err := dsl.ChainInfo(ctx)
	if err != nil {
		t.Fatalf("Failed to get chain info: %v", err)
	}

	// I'm not quite sure how to test other fields of ChainInfo, since they are dynamic
	if info.ChainID != chainId {
		t.Fatalf("Expected chain id %s, got %s", chainId, info.ChainID)
	}
}
