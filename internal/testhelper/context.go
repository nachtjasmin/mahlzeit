package testhelper

import (
	"context"
	"testing"
	"time"
)

// Context returns a context that is cancelled when either:
//   - the test is completed OR
//   - 10 seconds elapsed
//
// This ensures that no (sub-)test is running longer than 10 seconds.
func Context(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ctx, cancel = context.WithTimeout(ctx, time.Second*10)
	t.Cleanup(cancel)

	return ctx
}
