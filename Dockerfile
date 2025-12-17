# Build stage
FROM golang:1.24

# Install base packages
RUN apt-get update -qq && \
    apt-get install --no-install-recommends -y curl && \
    rm -rf /var/lib/apt/lists /var/cache/apt/archives


# Set working directory
WORKDIR /app


# Copy Go module files
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Install tailwind
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && \
    mv tailwindcss-linux-x64 ./bin/tailwindcss
# Build CSS with Tailwind
RUN ./bin/tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify

# Generate templ files
RUN templ generate

# Build the Go application
RUN go build -o /app/server ./cmd/server/main.go

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"]