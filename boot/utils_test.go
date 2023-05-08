package boot_test

import (
	"context"
	"testing"
)

func testContext(t *testing.T) context.Context {
	ctx := context.Background()
	var cancel func()
	if d, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, d)
	}
	if cancel == nil {
		ctx, cancel = context.WithCancel(ctx)
	}
	t.Cleanup(cancel)
	return ctx
}
