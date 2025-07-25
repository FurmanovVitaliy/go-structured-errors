# Structured Errors for Go
## Package `apperror`

This package provides a robust and structured error handling mechanism for Go applications, particularly tailored for microservices architectures using gRPC and grpc-gateway.

### Features

- **Structured Errors**: Create errors with a service name, a unique code, a human-readable message, and additional key-value fields for context.
- **Error Wrapping**: Supports standard Go error wrapping (`errors.Is`, `errors.As`) via the `Unwrap()` method.
- **Fluent API**: Chain methods to build and enrich errors declaratively (`err.WithField(...).WithGRPCCode(...)`).
- **gRPC Integration**: Seamlessly convert application errors to and from `gRPC status` objects, preserving all structured details.
- **grpc-gateway Ready**: Includes a custom error handler to translate gRPC errors into clean, structured JSON HTTP responses for your clients.
- **Traceability**: Directly integrates with OpenTelemetry by extracting the `trace_id` from the `SpanContext` in the `context`.
- **Structured Logging**: Provides a JSON marshaler to produce log-friendly output, automatically including the `trace_id`.

---

## How to Use

### 1. Install

Add the package to your project:

```bash
go get github.com/FurmanovVitaliy/go-structured-errors
```

### 2. Define Error Templates

Define reusable error templates for your application.

```go
import "github.com/FurmanovVitaliy/go-structured-errors"

var (
	ErrUserNotFound = apperror.New("user-service", "US-404", "user not found")
	ErrInternal     = apperror.New("database", "DB-500", "internal database error")
)
```

### 3. Add Context and Convert for gRPC

In your service logic, wrap low-level errors or use your templates. Add context with the fluent API and convert to a gRPC status before returning.

```go
import (
    "google.golang.org/grpc/codes"
    "your/project/apperror"
)

func someServiceLogic(id string) error {
    dbErr := findUserInDB(id) // some low-level call
    if dbErr != nil {
        appErr := apperror.Wrap(dbErr, apperror.ErrInternal).
            WithField("user_id", id).
            WithGRPCCode(codes.Internal)

        return appErr.GRPCStatus().Err()
    }
    return nil
}
```

---

## Running the Example

A complete, runnable example is available in the `_examples` directory.
To run it from the project's root directory, use the following command:

```sh
go run ./_examples
```

This will execute the demonstration and print a detailed output showcasing all features of the package.

### Example Output 

Running the example will produce the following output, demonstrating all features of the package.

```text
--- Demonstrating the apperror package ---

--- Scenario 1: Wrapped Database Error ---
Simple Error() string:
[database:DB-500] internal database error [fields:{}]: connection refused

--- Scenario 2: Business Logic Error with Fields ---
Simple Error() string:
[user-service:US-404] user not found [fields:{"attempt":"1", "searched_id":"user-456", "source":"web_api"}]

--- Scenario 3: gRPC Status Simulation ---
gRPC Status Code: NotFound
gRPC Status Message: user not found
gRPC Status Details: [service:"user-service" code:"US-404" message:"user not found" fields:<key:"attempt" value:"1" > fields:<key:"searched_id" value:"user-456" > fields:<key:"source" value:"web_api" > ]

--- Scenario 4: Recreating AppError from gRPC Status ---
Error recreated from status:
[user-service:US-404] user not found [fields:{"attempt":"1", "searched_id":"user-456", "source":"web_api"}]

--- Scenario 5: Structured JSON Logging with Trace ID ---
JSON output for logger:
{"service":"user-service","code":"US-404","message":"user not found","fields":{"attempt":"1","searched_id":"user-456","source":"web_api"},"trace_id":"a1b2c3d4e5f61234a1b2c3d4e5f61234"}

--- Scenario 6: Standard Library Error Unwrapping ---
✅ errors.Is check passed: wrappedErr contains the original dbErr.
✅ Unwrapped original cause: connection refused

--- Scenario 7: ToJSON with a standard error ---
JSON for a standard error: {"error":"file does not exist"}
``` 