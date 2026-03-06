# LazyLinear justfile

set dotenv-load

# Default recipe - show available commands
default:
    @just --list

# Variables
binary := "lazylinear"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
commit := `git rev-parse --short HEAD 2>/dev/null || echo "none"`
auth_pkg := "github.com/mane-pal/lazylinear/pkg/auth"
oauth_client_id := env_var_or_default("LINEAR_OAUTH_CLIENT_ID", "")
oauth_client_secret := env_var_or_default("LINEAR_OAUTH_CLIENT_SECRET", "")
ldflags := "-X main.version=" + version + " -X main.commit=" + commit + " -X " + auth_pkg + ".ClientID=" + oauth_client_id + " -X " + auth_pkg + ".ClientSecret=" + oauth_client_secret

# Build the binary
build:
    go build -ldflags "{{ldflags}}" -o {{binary}} ./cmd/lazylinear

# Build and run
run: build
    ./{{binary}}

# Run without building (faster iteration)
dev:
    go run ./cmd/lazylinear

# Run tests
test:
    go test ./...

# Run tests with verbose output
test-v:
    go test -v ./...

# Run tests for a specific package
test-pkg pkg:
    go test -v ./pkg/{{pkg}}/...

# Format code
fmt:
    go fmt ./...

# Tidy dependencies
tidy:
    go mod tidy

# Lint code (requires golangci-lint)
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -f {{binary}}

# Install locally
install: build
    cp {{binary}} ~/.local/bin/

# Watch for changes and rebuild (requires entr)
watch:
    find . -name "*.go" | entr -c just build

# Watch and run (requires entr)
watch-run:
    find . -name "*.go" | entr -c just dev

# Show project structure
tree:
    find . -type f -name "*.go" | head -30

# Count lines of code
loc:
    @find . -name "*.go" -exec wc -l {} + | tail -1
