package traceid

import (
	"context"
	"fmt"
	"time"
)

func New() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

type traceIdContextKey struct{}

func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIdContextKey{}, id)
}

func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(traceIdContextKey{}).(string)
	return id, ok
}
