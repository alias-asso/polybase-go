default:
    @just --list

# Run development server with air
dev:
    air

# Build both binaries
build:
    go build ./cmd/polybase-web
    go build ./cmd/polybase-cli

# Setup test environment
setup:
    sqlite3 polybase.db < migrations/001_init.sql

# Clean test data
clean:
    rm -rf .cache/
    rm -f polybase-web
    rm -f polybase-cli
