Build & Test

- Build binary (project root):
  go build ./...

- Run all tests:
  go test ./...

- Run a single test (package-level):
  go test ./yummy/recipe -run TestName

- Run a single test by fully-qualified package path:
  go test github.com/your/module/path/yummy/recipe -run TestName

- Quick vet/static checks:
  go vet ./...
  go list -f '{{.Dir}}' ./... | xargs -n1 go vet

- Formatting & imports:
  gofmt -w .
  goimports -w .   # if installed

- Linting (recommended):
  golangci-lint run  # if golangci-lint is installed/configured

Code style guidelines

- Formatting: always run gofmt (or gofmt via editor) and goimports before committing.
- Imports: group standard library first, blank line, then external modules. Keep imports minimal.
- Naming: Use camelCase for variables, PascalCase for exported identifiers. Avoid stuttering (package name repeated in type name).
- Error handling: return wrapped errors using fmt.Errorf("...: %w", err) when adding context. Check errors immediately after calls.
- Types: prefer small, focused structs. Use interfaces only for behaviour that will be mocked or have multiple implementations.
- Logging: use the project logger (yummy/log) if available; otherwise keep logs minimal and structured.
- Tests: keep tests hermetic, avoid network/file system side effects. Use t.Run for subtests. Name tests TestBehaviour_Condition or TestFunction_Scenario.
- Concurrency: prefer context.Context for cancellation and timeouts. Avoid sharing mutable state between goroutines.
- Comments: keep comments short and factual. Public types/functions must have doc comments starting with the identifier.
- Error messages: user-facing messages should be lower-case, no trailing punctuation. Internal logs may be more verbose.

Repo-specific notes

- No .cursor or Copilot instruction files were found; there are no special cursor/copilot rules to include.
- If you add environment variables, create a .env in the project root with placeholders and document them here.

If you want, I can add repository-specific lint config (golangci-lint) or run the test suite and lint checks now.
