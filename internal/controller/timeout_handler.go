package controller

import (
	"context"
	"time"
)

var (
	defaultTimeout = 5 * time.Second
)

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
