package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	pb "github.com/FurmanovVitaliy/grpc-api/gen/go/errors/errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// HTTPError - структура для JSON-ответа
type HTTPError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Details *pb.ErrorDetail `json:"details,omitempty"`
}

// GRPCAppErrorHandler - кастомный обработчик ошибок для grpc-gateway
func GRPCAppErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	// Попытка достать детали ошибки
	var errorDetail *pb.ErrorDetail
	for _, detail := range st.Details() {
		if d, ok := detail.(*pb.ErrorDetail); ok {
			errorDetail = d
			break
		}
	}

	// Маппинг gRPC-кода на HTTP-код
	httpCode := runtime.HTTPStatusFromCode(st.Code())

	// Формируем JSON-ответ
	resp := HTTPError{
		Code:    httpCode,
		Message: st.Message(),
		Details: errorDetail,
	}

	// Отправка ответа клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(resp)
}
