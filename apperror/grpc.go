package apperror

import (
	pb "github.com/FurmanovVitaliy/grpc-api/gen/go/errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WithGRPCCode returns a new AppError with the specified gRPC status code.
// It allows for setting the gRPC code that will be used when converting the error to a gRPC Status.
func (e *AppError) WithGRPCCode(code codes.Code) *AppError {
	copy := *e
	copy.grpcCode = int(code)
	return &copy
}

// GRPCStatus converts the AppError to a gRPC *status.Status.
// It embeds the AppError's details into the status, allowing them to be
// extracted by a gRPC client or gateway.
// If no gRPC code was set, it defaults to codes.Unknown.
func (e *AppError) GRPCStatus() *status.Status {
	code := codes.Code(e.grpcCode)
	// An error should not have a status of OK (0).
	// If no specific gRPC code was set, default to Unknown.
	// This handles cases where an error is created via apperror.New()
	// without a subsequent call to .WithGRPCCode().
	if code == codes.OK {
		code = codes.Unknown
	}

	st := status.New(code, e.Message)
	detail := &pb.ErrorDetail{
		Service: e.Service,
		Code:    e.Code,
		Message: e.Message,
		Fields:  e.Fields,
	}
	stWithDetail, err := st.WithDetails(detail)
	if err != nil {
		return status.New(codes.Internal, "failed to marshal error")
	}
	return stWithDetail
}

// FromGRPCStatus creates an AppError from a gRPC *status.Status.
// It attempts to extract the detailed error information from the status's details.
// If no details are found, it creates a new AppError from the status's message and code.
func FromGRPCStatus(st *status.Status) *AppError {
	if st == nil {
		return nil
	}

	for _, d := range st.Details() {
		if detail, ok := d.(*pb.ErrorDetail); ok {
			return &AppError{
				Service:  detail.Service,
				Code:     detail.Code,
				Message:  detail.Message,
				Fields:   detail.Fields,
				grpcCode: int(st.Code()),
			}
		}
	}

	// Fallback for a standard gRPC error without custom details.
	return &AppError{
		Service:  "unknown",
		Code:     "00000",
		Message:  st.Message(),
		grpcCode: int(st.Code()),
	}
}
