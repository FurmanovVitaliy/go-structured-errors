package apperror

import (
	"context"
	"encoding/json"
	"net/http"

	pb "github.com/FurmanovVitaliy/grpc-api/gen/go/errors/errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// traceIDKey is the type used for the trace_id key in the context.
// Using a custom type prevents key collisions.
type TraceIDKey struct{}

// HTTPError represents a structured error response for an HTTP API.
type HTTPError struct {
	Code        int               `json:"code"`
	Service     string            `json:"service"`
	ServiceCode string            `json:"service_code"`
	Message     string            `json:"message"`
	Fields      map[string]string `json:"fields,omitempty"`
	TraceID     string            `json:"trace_id,omitempty"`
}

// GRPCAppErrorHandler is a custom error handler for grpc-gateway.
// It translates gRPC errors into a structured JSON HTTP response.
// It extracts the AppError details from the gRPC status and includes a trace ID from the context.
func GRPCAppErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	st := status.Convert(err)
	httpCode := runtime.HTTPStatusFromCode(st.Code())
	resp := HTTPError{
		Code:    httpCode,
		Message: st.Message(),
	}

	// Check for custom error details in the gRPC status.
	for _, d := range st.Details() {
		if detail, ok := d.(*pb.ErrorDetail); ok {
			resp.Service = detail.Service
			resp.ServiceCode = detail.Code
			// If there's a custom message in details, prefer it over the gRPC status message.
			if detail.Message != "" {
				resp.Message = detail.Message
			}
			resp.Fields = detail.Fields
			break
		}
	}

	// Add trace_id to the response if it exists in the context.
	if traceID, ok := ctx.Value(TraceIDKey{}).(string); ok {
		resp.TraceID = traceID
	}

	w.Header().Set("Content-Type", marshaler.ContentType(resp))
	w.WriteHeader(httpCode)
	jsonErr := json.NewEncoder(w).Encode(resp)
	if jsonErr != nil {
		// If encoding fails, fallback to the default error handler.
		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
	}
}
