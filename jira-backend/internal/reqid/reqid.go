// Package reqid carries a per-request correlation id through the backend:
// HTTP middleware -> context -> outgoing gRPC metadata. The connector reads it
// back from the incoming metadata so a single request can be traced across
// REST -> gRPC -> Kafka by the "request_id" log field.
package reqid

import (
	"context"

	"github.com/google/uuid"
)

const HeaderName = "X-Request-Id"

const MetadataKey = "x-request-id"

type ctxKey struct{}

func New() string {
	return uuid.NewString()
}

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}
