// Package apperror provides a structured error handling mechanism for Go applications.
// It allows for creating errors with service context, error codes, and additional fields,
// supports error wrapping, and is designed to work seamlessly with gRPC and structured logging.
package apperror

import (
	"fmt"
)

// ErrorFields is a type for structured error fields.
type ErrorFields map[string]string

// AppError is a custom error type that includes service information,
// a unique code, a human-readable message, and additional structured fields.
// It supports error wrapping and can be converted to a gRPC status.
type AppError struct {
	Service  string      `json:"service,omitempty"` // example: user-service
	Code     string      `json:"code,omitempty"`    // example: 001
	Message  string      `json:"message"`           // human readable message
	Fields   ErrorFields `json:"fields,omitempty"`  // additional fields, key-value pairs
	cause    error       // wrapped error
	grpcCode uint32      // used in grpc.go
	traceID  string      // set in marshal.go
}

// New creates a new AppError.
// It serves as the root error in a chain.
func New(service string, code string, message string) *AppError {
	return &AppError{
		Service: service,
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with an AppError, creating a chain.
// If the original error is nil, Wrap returns nil.
func Wrap(err error, appErr *AppError) *AppError {
	if err == nil {
		return nil
	}
	copyErr := *appErr
	copyErr.cause = err
	return &copyErr
}

// WithField adds a single key-value pair to the error's fields.
// It returns a new AppError to maintain immutability.
func (e *AppError) WithField(key, value string) *AppError {
	return e.AddFields(ErrorFields{key: value})
}

// AddFields adds multiple key-value pairs to the error's fields.
// It creates a new AppError to maintain immutability.
// If a key already exists, its value is overwritten.
func (e *AppError) AddFields(fields ErrorFields) *AppError {
	if len(fields) == 0 {
		return e
	}
	copyErr := *e
	newFields := make(ErrorFields, len(e.Fields)+len(fields))

	for k, v := range e.Fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	copyErr.Fields = newFields
	return &copyErr
}

// WithFields replaces the error's fields with the provided ones.
// It returns a new AppError to maintain immutability.
func (e *AppError) WithFields(fields ErrorFields) *AppError {
	copyErr := *e
	copyErr.Fields = fields
	return &copyErr
}

// Error returns a string representation of the error, suitable for logging.
// The format includes service, code, message, fields, and the wrapped error.
func (e *AppError) Error() string {
	// Format the error message with a service-code prefix for clear logging.
	msg := fmt.Sprintf("[%s:%s] %s", e.Service, e.Code, e.Message)

	if len(e.Fields) > 0 {
		// Add fields for additional context.
		// NOTE: map iteration order is not guaranteed.
		msg += " [fields:{"
		first := true
		for k, v := range e.Fields {
			if !first {
				msg += ", "
			}
			msg += fmt.Sprintf("%q:%q", k, v)
			first = false
		}
		msg += "}]"
	}

	if e.cause != nil {
		// Append the wrapped error.
		msg += ": " + e.cause.Error()
	}
	return msg
}

// Unwrap returns the wrapped error, to support errors.Is and errors.As.
func (e *AppError) Unwrap() error {
	return e.cause
}

// Is checks if the target error is an AppError with the same identity.
// It allows AppError to be used with errors.Is.
// The comparison is based on Service and Code. If the target error also has fields,
// it checks if all of those fields are present and have the same values in the source error.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}

	// Core identity check
	if e.Service != t.Service || e.Code != t.Code {
		return false
	}

	// If the target has fields, ensure they are all present in the source error.
	// This allows for flexible testing: you can check for a subset of fields.
	for key, val := range t.Fields {
		if v, ok := e.Fields[key]; !ok || v != val {
			return false
		}
	}

	return true
}
