# QR Code Maintenance - Build Commands

.PHONY: tailwind tailwind-watch tailwind-build server

# Build Tailwind CSS once
tailwind:
	pnpm tailwindcss -i static/css/input.css -o static/css/output.css

# Watch Tailwind CSS for changes during development
tailwind-watch:
	pnpm tailwindcss -i static/css/input.css -o static/css/output.css --minify --watch

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

# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
live/templ:
	templ generate --watch -v

# run air to detect any go file changes to re-build and re-run the server.
live/server:
	air

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	pnpm tailwindcss -i static/css/input.css -o static/css/output.css --minify --watch

# start all 5 watch processes in parallel.
live: 
	make -j3 live/templ live/server live/tailwind