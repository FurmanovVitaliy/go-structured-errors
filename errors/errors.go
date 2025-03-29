package errors

import (
	"fmt"

	pb "github.com/FurmanovVitaliy/grpc-api/gen/go/errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type ErrorFields map[string]string

// AppError - базовая структура ошибки
type AppError struct {
	Service string      // Идентификатор сервиса
	Code    string      // Уникальный код ошибки
	Message string      // Человекочитаемое сообщение
	Fields  ErrorFields // Дополнительные поля
	Cause   error       // Исходная ошибка

	grpcCode codes.Code // Внутреннее поле для gRPC кода
}

// GRPCError - интерфейс для ошибок с gRPC метаданными
type GRPCError interface {
	error
	GRPCStatus() *status.Status
	WithGRPCCode(codes.Code) *AppError
}

func (e *AppError) Error() string {
	msg := fmt.Sprintf("[%s-%s] %s", e.Service, e.Code, e.Message)
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}
	if len(e.Fields) > 0 {
		msg += " | Fields: " + fmt.Sprintf("%+v", e.Fields)
	}
	return msg
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// Добавляет gRPC код к ошибке
func (e *AppError) WithGRPCCode(c codes.Code) *AppError {
	e.grpcCode = c
	return e
}

// Реализует интерфейс для gRPC статуса
func (e *AppError) GRPCStatus() *status.Status {
	code := codes.Unknown
	if e.grpcCode != codes.OK {
		code = e.grpcCode
	}

	st := status.New(code, e.Message)
	errDetail := &pb.ErrorDetail{
		Service: e.Service,
		Code:    e.Code,
		Message: e.Message,
		Fields:  e.Fields,
	}

	st, err := st.WithDetails(errDetail)
	if err != nil {
		return status.New(codes.Internal, "failed to add error details")
	}
	return st
}

// Создание новой ошибки
func New(service, code, message string) *AppError {
	return &AppError{
		Service: service,
		Code:    code,
		Message: message,
		Fields:  make(ErrorFields),
	}
}

// Обертка для ошибок
func Wrap(err error, service, code, message string) *AppError {
	return &AppError{
		Service: service,
		Code:    code,
		Message: message,
		Cause:   err,
		Fields:  make(ErrorFields),
	}
}

// Добавление полей
func (e *AppError) WithFields(fields ErrorFields) *AppError {
	for k, v := range fields {
		e.Fields[k] = v
	}
	return e
}

// Конвертация в JSON
func (e *AppError) MarshalJSON() ([]byte, error) {
	return protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(e.ToProto())
}

func (e *AppError) ToProto() *pb.ErrorDetail {
	return &pb.ErrorDetail{
		Service: e.Service,
		Code:    e.Code,
		Message: e.Message,
		Fields:  e.Fields,
	}
}

// Парсинг из gRPC статуса
func FromGRPCStatus(st *status.Status) *AppError {
	if st == nil {
		return nil
	}

	code := st.Code()
	details := st.Details()

	for _, detail := range details {
		if errDetail, ok := detail.(*pb.ErrorDetail); ok {
			return &AppError{
				Service:  errDetail.Service,
				Code:     errDetail.Code,
				Message:  errDetail.Message,
				Fields:   ErrorFields(errDetail.Fields),
				grpcCode: code,
			}
		}
	}

	return &AppError{
		Service:  "unknown",
		Code:     "00000",
		Message:  st.Message(),
		grpcCode: code,
	}
}

// Примеры использования
var (
	ErrInternal     = New("common", "00100", "internal error").WithGRPCCode(codes.Internal)
	ErrInvalidInput = New("common", "00101", "invalid input").WithGRPCCode(codes.InvalidArgument)
)
