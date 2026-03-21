# AGENTS.md

This file provides guidelines for agentic coding assistants working in this repository.

## Build, Lint, and Test Commands

### CI
Using GitHub Actions.

- Use GitHub Actions for CI/CD
- Keep workflows in `.github/workflows/`
- Use `actions/checkout@v6` version >=6 for checking out code
- Use `actions/setup-go@v6` version >=6 for setting up Go
- Use `planetscale/ghcommit-action@v0.2.20` version >=0.2.20 for committing changes

### Build
```bash
# Build the executable
go build -o bin/awesome-repositories

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/awesome-repositories
GOOS=darwin GOARCH=amd64 go build -o bin/awesome-repositories
```

### Run
```bash
# Run directly
go run main.go -username <github-username> [options]

# Run with options
go run main.go -username <github-username> -page-size 100 -limit 1000

# Run the built binary
./bin/awesome-repositories -username <github-username>
```

### Test
```bash
# Run all tests (no tests exist yet, but add tests as needed)
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run a single test file
go test -v ./github -run TestFunctionName

# Run a specific test
go test -v ./... -run ^TestSpecificFunction$
```

### Lint
```bash
# Format code
go fmt ./...

# Run vet (built-in Go linter)
go vet ./...

# Install and run golangci-lint if desired
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run
```

### Dependencies
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Code Style Guidelines

### Imports
- Use standard library imports first, then third-party packages
- Group imports into three blocks with blank lines between them
- Use short package aliases where appropriate (e.g., `u "github.com/eagleusb/awesome-repositories/utils"`)
- Keep imports sorted alphabetically within each group

Example:
```go
import (
    "bufio"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/google/go-github/v83/github"
)

import (
    u "github.com/eagleusb/awesome-repositories/utils"
)
```

### Formatting
- Use `gofmt` for consistent formatting (standard Go style)
- Run `go fmt ./...` before committing
- Use tabs for indentation (Go standard)
- Keep line length reasonable (no strict limit, but prefer <120 chars)
- Use `defer` for cleanup operations

### Types
- Export types (PascalCase) for external use
- Unexport types (camelCase) for internal use
- Use struct composition instead of inheritance
- Embed types when composition makes sense

Example:
```go
// Exported type
type GitHubClient struct {
    client *github.Client
    config *githubConfig
}

// Internal type
type githubConfig struct {
    username     string
    itemsPerPage int
    maxItems     int
}
```

### Naming Conventions
- **Packages**: lowercase, single word (e.g., `github`, `utils`)
- **Constants**: UPPERCASE_SNAKE_CASE for exported, lowercaseSnakeCase for internal
- **Variables**: camelCase
- **Functions**: PascalCase for exported, camelCase for internal
- **Interfaces**: PascalCase ending with "er" suffix when applicable
- **Acronyms**: Keep capitalization (e.g., `GITHUB_TOKEN`, `HTTPClient`)

Examples:
```go
// Constants
const (
    DefaultTimeout = 15 * time.Second
    maxRetries     = 3
)

// Functions
func NewGitHubClient() (*GitHubClient, error) { ... }
func (c *GitHubClient) GetStarredRepos(...) { ... }
func internalHelper() { ... }

// Variables
var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
```

### Error Handling
- Always handle errors explicitly
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Return errors as last return value
- Check errors immediately after function calls
- Provide context in error messages

Examples:
```go
// Good - wrap errors with context
file, err := os.Create(filename)
if err != nil {
    return fmt.Errorf("failed to create file %s: %w", filename, err)
}

// Good - nil checks
if r == nil {
    fmt.Printf("Skipping nil repository\n")
    continue
}

// Good - validation
if GITHUB_TOKEN == "" {
    return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
}
```

### Context and Timeouts
- Use `context.Context` for cancelable operations
- Always set timeouts for network operations
- Use `defer cancel()` to ensure context cleanup

Example:
```go
ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
defer cancel()

iter := c.client.Activity.ListStarredIter(ctx, username, opts)
```

### Resource Management
- Use `defer` for cleanup (file.Close(), writer.Flush(), cancel())
- Close resources in reverse order of opening
- Check errors from deferred operations when critical

Example:
```go
file, err := os.Create(filename)
if err != nil {
    return err
}
defer file.Close()

writer := bufio.NewWriter(file)
defer writer.Flush()

// Use resources...
```

### Slices and Maps
- Use `make()` for slices and maps when size is known
- Use `slices.SortFunc` for custom sorting (Go 1.25+)
- Initialize maps in constructors
- Check for nil before dereferencing

Example:
```go
// Initialize in constructor
Repos: &githubRepos{
    ByLanguage: make(map[string][]*githubRepo),
    ByCategory: make(map[string][]*githubRepo),
},

// Sort with custom function
slices.SortFunc(sortedRepos, func(a, b *githubRepo) int {
    return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
})
```

### Package Structure
- `main.go`: Application entry point and CLI flags
- `github/`: GitHub API client and data handling
- `utils/`: Utility functions (string sanitization, directory operations)
- Keep packages focused and cohesive

### Environment Variables
- Use uppercase for environment variable names
- Check for empty values and provide clear error messages
- Required environment variables should be validated early

Example:
```go
var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")

func NewGitHubClient() (*GitHubClient, error) {
    if GITHUB_TOKEN == "" {
        return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
    }
    // ...
}
```

### Configuration
- Use flag package for CLI arguments
- Provide sensible defaults
- Validate user input

Example:
```go
var (
    help     = flag.Bool("help", false, "show help")
    username = flag.String("username", "", "GitHub username")
    page     = flag.Int("page-size", 100, "number of repositories to fetch per page")
    limit    = flag.Int("limit", 1000, "maximum number of repositories to fetch")
)
```

### Logging and Output
- Use `fmt.Printf` for user-facing output (progress, status)
- Log important milestones during long-running operations
- Keep error messages descriptive and actionable

### Documentation
- Add package-level comments for exported packages
- Document exported functions and types
- Keep comments concise and accurate

## Notes for AI Agents
- This is a Go 1.25+ project
- The application fetches starred repositories and classifies them by language
