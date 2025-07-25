package apperror

import (
	"context"
	"encoding/json"
)

// ToJSON marshals an AppError to a JSON byte slice, suitable for structured logging.
// It enriches the error with a trace_id from the context, if available.
// The marshaling is done using a custom alias to avoid marshaling recursion.
func ToJSON(ctx context.Context, err error) []byte {
	appErr, ok := err.(*AppError)
	if !ok {
		// If the error is not an AppError, marshal it as a simple string.
		return []byte(`{"error":"` + err.Error() + `"}`)
	}

	// The use of an alias is a common Go pattern to avoid MarshalJSON recursion.
	// We create a new type `alias` which has the same structure as AppError
	// but none of its methods (including MarshalJSON).
	type alias AppError

	// Enrich the error with the trace_id from the context before marshaling.
	if traceID, ok := ctx.Value(TraceIDKey{}).(string); ok {
		appErr.traceID = traceID
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
