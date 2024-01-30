package datasource

import (
	"context"
	"fmt"
	"testing"
)

func TestStreamingAPI(t *testing.T) {
	data := []int{0, 1, 2, 3, 4, 5, 6, 8}
	ctx := context.TODO()

	newData := ToStream(ctx, data).
		Filter(ctx, func(x int) bool { return x%2 == 0 }).
		Map(ctx, func(x int) int { return x * 2 }).
		Collect(ctx)

	fmt.Println(newData)
}
