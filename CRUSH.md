# CRUSH.md

## Build, Lint, and Test Commands
- Build the project: `go build ./...`
- Run tests: `go test ./...`
- Run a single test: `go test -run TestFunctionName ./path/to/package`
- Lint the code: `golint ./...`

## Code Style Guidelines

### Imports
- Group imports by standard library, external packages, and local packages.
- Use all relevant imports, avoiding unnecessary ones.

### Formatting
- Adhere to Go standard formatting (use `gofmt` or `go fmt`).
- Maintain a consistent style across all files.

### Types
- Prefer the use of interface types where possible for flexibility.
- Keep types and functions small and focused.

### Naming Conventions
- Use camelCase for variables and functions, and PascalCase for types.
- Ensure names are descriptive and convey intent.

### Error Handling
- Always handle errors returned by functions.
- Use `fmt.Errorf` for wrapping errors with additional context.

### Comments
- Write comments for complex logic, explaining why something is done rather than what.
- Avoid comments that state the obvious.

## Cursor and Copilot Rules
_No rules found._  

# Additional Notes
- Ensure to keep dependencies updated to minimize security vulnerabilities.
- Follow best practices for concurrent programming when dealing with goroutines and channels.
