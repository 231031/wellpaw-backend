package controller

import (
	"context"
	"time"
)

var (
	defaultTimeout = 15 * time.Second
)

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
