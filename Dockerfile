# Build stage
FROM golang:1.24

# Install base packages
RUN apt-get update -qq && \
    apt-get install --no-install-recommends -y curl && \
    rm -rf /var/lib/apt/lists /var/cache/apt/archives

# Set working directory
WORKDIR /app

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g corepack && \
    corepack enable pnpm && \
    pnpm i tailwindcss@latest @tailwindcss/cli@latest daisyui@latest

# Copy Go module files
COPY go.mod go.sum ./

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Download Go dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Install tailwind
RUN 
# Build CSS with Tailwind
RUN pnpm exec tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify

# Generate templ files
RUN templ generate

# Build the Go application
RUN go build -o /app/server ./cmd/server/main.go

# Build the Database command
RUN go build -o /app/database ./cmd/database/main.go

# Build the River command
RUN go build -o /app/background_jobs ./cmd/background_jobs/main.go

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"]