# QR Code Portal Maintenance System

A web application for managing portal maintenance through QR codes. Each portal has a unique QR code that provides different functionality based on who scans it.

## 🎯 Purpose

This tool enables efficient portal maintenance management by providing:

- **Public Access**: Citizens can check maintenance history and request services
- **Maintenance Teams**: Admin interface for recording maintenance activities and managing portals

## 🏗️ Tech Stack

- **Backend**: Go with Echo web framework
- **Frontend**: Templ templates with HTMX for interactivity  
- **Styling**: Tailwind CSS
- **Database**: PostgreSQL with sqlx
- **Development**: Dev containers with Docker

## 📋 Prerequisites

- Go 1.24.5+
- Node.js & pnpm (for Tailwind CSS)
- PostgreSQL (via dev container)
- Docker (for dev container setup)

## 🚀 Getting Started

### 1. Database Setup
```bash
# Create database
go run cmd/setup/main.go

# Run migrations
go run cmd/migrate/main.go

# Seed with sample data
go run cmd/seed/main.go
```

### 2. Frontend Setup
```bash
# Install CSS dependencies  
pnpm install

# Generate CSS
make tailwind

# Or watch for changes during development
make tailwind-watch
```

### 3. Generate Templates
```bash
# Generate Go code from templ templates
templ generate
```

### 4. Start Development Server
```bash
# Start server with template generation
make server

# Or manually
go run cmd/server/main.go
```

## 🔄 User Scenarios

### Public Users
- Scan QR code on portal
- View last maintenance date
- See responsible maintenance company
- Request maintenance if needed

### Maintenance Teams  
- Scan QR code for admin access
- Record new maintenance activities
- Update portal status and information
- Manage maintenance history

## 🧪 Testing

The project includes comprehensive tests:

- **Unit tests**: Handler logic and template rendering
- **Integration tests**: Server routing and 404 handling  
- **Template tests**: HTML output and accessibility

```bash
# Run all tests
make test

# Run with detailed output
make test-verbose

# Run specific package tests
go test ./internal/handlers
go test ./internal/templates  
go test ./cmd/server
```

## 🐳 Development Container

This project uses VS Code dev containers with:
- Go development environment
- PostgreSQL database with persistent storage
- Node.js for frontend tooling
- Pre-configured extensions and settings

---

Built with ❤️ for efficient portal maintenance management