package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	pb "github.com/FurmanovVitaliy/grpc-api/gen/go/errors/errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTPError struct for custom error response
type HTTPError struct {
	Code        int               `json:"code"`
	Service     string            `json:"service,omitempty"`
	ServiceCode string            `json:"service_code,omitempty"`
	Message     string            `json:"message"`
	Fields      map[string]string `json:"fields,omitempty"`
}

// / GRPCAppErrorHandler - handler for custom error handling
func GRPCAppErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	httpCode := runtime.HTTPStatusFromCode(st.Code())

	var errorDetail *pb.ErrorDetail
	for _, detail := range st.Details() {
		if d, ok := detail.(*pb.ErrorDetail); ok {
			errorDetail = d
			break
		}
	}

	if st.Code() == codes.Unavailable && errorDetail == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)

		resp := HTTPError{
			Code:    http.StatusServiceUnavailable,
			Message: "Service unavailable. Please try again later.",
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// Обычная обработка ошибок с деталями
	resp := HTTPError{
		Code:    httpCode,
		Message: st.Message(),
	}

	if errorDetail != nil {
		resp.Service = errorDetail.Service
		resp.ServiceCode = errorDetail.Code
		resp.Fields = errorDetail.Fields
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(resp)
}
