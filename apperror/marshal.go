package apperror

import (
	"context"
	"encoding/json"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

// ToJSON marshals an AppError to a JSON byte slice, suitable for structured logging.
// It enriches the error with a trace_id from the context, if available.
// The marshaling is done using a custom alias to avoid marshaling recursion.
func ToJSON(ctx context.Context, err error) []byte {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		// If the error is not an AppError, marshal it as a simple string.
		return []byte(`{"error":"` + err.Error() + `"}`)
	}

	// The use of an alias is a common Go pattern to avoid MarshalJSON recursion.
	// We create a new type `alias` which has the same structure as AppError
	// but none of its methods (including MarshalJSON).
	type alias AppError

	// Enrich the error with the trace_id from the context before marshaling.
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		appErr.traceID = span.TraceID().String()
	}

	b, marshalErr := json.Marshal(&struct {
		*alias

		TraceID string `json:"trace_id,omitempty"`
	}{
		alias:   (*alias)(appErr),
		TraceID: appErr.traceID,
	})
	if marshalErr != nil {
		return []byte(`{"error":"failed to marshal app error: ` + marshalErr.Error() + `"}`)
	}

	return b
}
