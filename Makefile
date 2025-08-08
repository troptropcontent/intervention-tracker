# QR Code Maintenance - Build Commands

.PHONY: tailwind tailwind-watch tailwind-build server

# Build Tailwind CSS once
tailwind:
	pnpm tailwindcss -i static/css/input.css -o static/css/output.css

# Watch Tailwind CSS for changes during development
tailwind-watch:
	pnpm tailwindcss -i static/css/input.css -o static/css/output.css --watch

# Build Tailwind CSS for production (minified)
tailwind-build:
	pnpm tailwindcss -i static/css/input.css -o static/css/output.css --minify

# Start the server
server:
	templ generate && go run cmd/server/main.go

# Run tests (clean output)
test:
	go test ./... | grep -v "\[no test files\]"

# Run tests with verbose output (clean)
test-verbose:
	go test ./... -v | grep -v "\[no test files\]"