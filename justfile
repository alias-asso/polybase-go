# Run development server with air
dev:
  air

auth-server:
  glauth -c glauth.cfg

# Build both binaries
build: build-server build-cli

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
  tailwindcss -i static/css/tailwind.css -o static/css/styles.css --minify
  templ generate
  go build -o target/polybased ./polybased

# Build cli binaries
build-cli:
  mkdir -p target
  go build -o target/polybase ./polybase
