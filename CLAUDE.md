# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a QR code maintenance project set up as a Go development environment with PostgreSQL database support. The repository is currently minimal, containing only development container configuration files.

## Development Environment

This project uses a dev container setup with:
- **Language**: Go 1.24
- **Database**: PostgreSQL (latest)
- **Container**: Based on Microsoft's Go dev container template
- **Additional Features**: Node.js support included

## Container Setup

The dev container is configured with:
- Go development environment in `/workspaces/qr_code_maintenance`
- PostgreSQL database service with persistent volume storage
- Network sharing between app and database containers
- Environment variables loaded from `.devcontainer/.env`

## Current State

The repository appears to be in early setup phase with only dev container configuration present. No source code, build scripts, or package files have been created yet.

## Next Steps for Development

When source code is added, typical Go project commands will likely include:
- `go mod init` - Initialize Go module
- `go build` - Build the application  
- `go test ./...` - Run tests
- `go run main.go` - Run the application

Database connection will be available on the internal docker network to the PostgreSQL service.