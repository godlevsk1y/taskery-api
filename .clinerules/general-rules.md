# Go Backend Development Rules

You are a Go Backend developer.

Follow these rules for all Go-related tasks:

1. **Architecture & Style**: 
   - Always follow basic Go v1.26 style and coding conventions
   - Always match the code style of existing files in the same directory.
   - Use `handlers.DecodeAndValidate` for request parsing.
   - Use `handlers.WriteJSON` for success and `handlers.WriteError` for errors.
   - Use `slog` for logging and `validator/v10` for DTOs.

2. **HTTP Handlers**:
   - Wrap service functions into interfaces (like Deleter in the @/internal/infrastructure/transport/http/v1/handlers/task/delete.go) for mocking.
   - Handlers must be structs with dependencies: Interface, Timeout, *slog.Logger, *validator.Validate.
   - Map service errors to HTTP status codes directly inside the handler (refer to @/internal/infrastructure/transport/http/v1/handlers/task/create.go for mapping style).

3. **Testing**:
   - Use table-driven tests.
   - Dependencies: `testify/require`, `gofakeit`, `httptest`.
   - Always define mock setup inside the test case struct.
   - Follow the pattern from @/internal/infrastructure/transport/http/v1/handlers/task/update_test.go.

4. **Automation**:
   - After defining or updating an interface, automatically run the `mockery` command to refresh mocks. Here is one and only correct syntax for calling mockery: "mockery". Run this program with no flags and arguments.