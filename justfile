default:
  @just --list

# Run development server with air
dev:
  air

# Build both binaries
build:
  mkdir -p target
  tailwindcss -i static/css/tailwind.css -o static/css/styles.css --minify
  templ generate
  go build -o target/polybase-http ./polybase-http
  go build -o target/polybase ./polybase

# Setup test environment
setup:
  sqlite3 polybase.db < migrations/001_init.sql

# Clean test data
clean:
  rm -fr .cache/
  rm -fr target/
