# Run development server with air
dev:
  hivemind

# Build both binaries
build: clean build-server build-cli

test:
  go test -cover ./...

# Setup test environment
setup:
  sqlite3 polybase.db < migrations/001_init.sql

# Clean test data
clean:
  rm -fr .cache/
  rm -fr target/

# Build server binaries
build-server:
  mkdir -p target
  tailwindcss -i static/css/main.css -o static/css/styles.css -m
  templ generate
  go build -o target/polybased ./polybased
  scdoc < polybased.1.scd | sed "s/1980-01-01/$(date '+%B %Y')/" > target/polybased.1

# Build cli binaries
build-cli:
  mkdir -p target
  go build -o target/polybase ./polybase
  scdoc < polybase.1.scd | sed "s/1980-01-01/$(date '+%B %Y')/" > target/polybase.1
