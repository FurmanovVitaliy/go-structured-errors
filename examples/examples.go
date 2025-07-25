package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/FurmanovVitaliy/go-structured-errors/apperror"
	"google.golang.org/grpc/codes"
)

// --- 1. Define reusable error templates ---
// This is a common practice in applications to ensure consistency.
var (
	ErrUserNotFound = apperror.New("user-service", "US-404", "user not found")
	ErrDatabase     = apperror.New("database", "DB-500", "internal database error")
)

// --- Let's simulate a function that interacts with a low-level system ---
func findUserInDB(id string) error {
	// Let's pretend the database returned a standard error
	return errors.New("connection refused")
}

func main() {
	fmt.Println("--- Demonstrating the apperror package ---")

	// --- 2. Wrapping a low-level error ---
	// Imagine we're in a repository layer. We get a generic error from the DB.
	dbErr := findUserInDB("123")

	// We wrap it with our structured AppError to add business context.
	wrappedErr := apperror.Wrap(dbErr, ErrDatabase)

	fmt.Println("\n--- Scenario 1: Wrapped Database Error ---")
	fmt.Printf("Simple Error() string:\n%s\n", wrappedErr)

	// --- 3. Creating a specific business logic error and adding context ---
	// This error doesn't wrap another one; it's born in our business logic.
	// We use the fluent API to add structured fields.
	businessErr := ErrUserNotFound.
		WithField("searched_id", "user-456").
		AddFields(apperror.ErrorFields{"attempt": "1", "source": "web_api"})

	fmt.Println("\n--- Scenario 2: Business Logic Error with Fields ---")
	fmt.Printf("Simple Error() string:\n%s\n", businessErr)

	// --- 4. Simulating gRPC transport ---
	// In a real gRPC service, you would convert your AppError to a gRPC Status.
	grpcStatus := businessErr.WithGRPCCode(codes.NotFound).GRPCStatus()

	fmt.Println("\n--- Scenario 3: gRPC Status Simulation ---")
	fmt.Printf("gRPC Status Code: %s\n", grpcStatus.Code())
	fmt.Printf("gRPC Status Message: %s\n", grpcStatus.Message())
	fmt.Printf("gRPC Status Details: %+v\n", grpcStatus.Details())

	// Now, let's simulate the other side (a gRPC client or gateway) receiving this status
	// and converting it back to an AppError.
	errFromStatus := apperror.FromGRPCStatus(grpcStatus)
	fmt.Println("\n--- Scenario 4: Recreating AppError from gRPC Status ---")
	fmt.Printf("Error recreated from status:\n%s\n", errFromStatus)

	// --- 5. Simulating logging with Trace ID ---
	// First, let's create a context with a trace ID.
	// In a real app, a middleware would do this.
	traceID := "a1b2c3d4-e5f6-g7h8-i9j0"
	ctx := context.WithValue(context.Background(), apperror.TraceIDKey{}, traceID)

	// Now, we'll convert our business error to JSON for logging.
	// The ToJSON function will automatically pick up the trace_id from the context.
	logJSON := apperror.ToJSON(ctx, businessErr)

	fmt.Println("\n--- Scenario 5: Structured JSON Logging with Trace ID ---")
	fmt.Printf("JSON output for logger:\n%s\n", string(logJSON))

	// --- 6. Demonstrating standard library unwrapping ---
	// This shows that your errors are compatible with `errors.Is` and `errors.Unwrap`.
	fmt.Println("\n--- Scenario 6: Standard Library Error Unwrapping ---")

	// Check if our wrappedErr is an ErrDatabase. This will be false because Wrap creates a copy.
	// To make this work, we'd need to compare fields or use a different pattern.
	// However, we can check for the original error.
	if errors.Is(wrappedErr, dbErr) {
		fmt.Println("✅ errors.Is check passed: wrappedErr contains the original dbErr.")
	} else {
		fmt.Println("❌ errors.Is check failed.")
	}

	// Let's get the original cause
	originalCause := errors.Unwrap(wrappedErr)
	if originalCause != nil {
		fmt.Printf("✅ Unwrapped original cause: %s\n", originalCause.Error())
	}

	// --- 7. Demonstration with a different error type for ToJSON ---
	fmt.Println("\n--- Scenario 7: ToJSON with a standard error ---")
	standardErrJSON := apperror.ToJSON(context.Background(), os.ErrNotExist)
	fmt.Printf("JSON for a standard error: %s\n", string(standardErrJSON))
}