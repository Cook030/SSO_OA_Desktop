// Package audit provides request-scoped provenance for database writes.
package audit

import "context"

type contextKey uint8

const (
	operatorIDKey contextKey = iota
	requestIDKey
)

// WithRequestID adds the request trace ID to a context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithOperatorID adds the authenticated operator ID to a context.
func WithOperatorID(ctx context.Context, operatorID uint64) context.Context {
	if operatorID == 0 {
		return ctx
	}
	return context.WithValue(ctx, operatorIDKey, operatorID)
}

// Provenance is the write attribution extracted from a request context.
type Provenance struct {
	OperatorID uint64
	RequestID  string
}

// FromContext extracts write attribution. Zero values represent system jobs or
// requests without an authenticated operator.
func FromContext(ctx context.Context) Provenance {
	if ctx == nil {
		return Provenance{}
	}

	provenance := Provenance{}
	if operatorID, ok := ctx.Value(operatorIDKey).(uint64); ok {
		provenance.OperatorID = operatorID
	}
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		provenance.RequestID = requestID
	}
	return provenance
}
