package demo

import (
	"context"
	"testing"
)

func Test_Context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-ctx.Done():
			t.Logf("%s\n", "cancel context")
			return
		}
	}()
}
