# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a QR code maintenance project - a full-stack Go web application for managing portals and QR codes with PostgreSQL database support.

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

## Architecture & Technology Choices

**Application Type**: Server-side rendered web application (not SPA/API)
**Frontend**: Go templates (templ) with HTMX and STIMULUS for interactivity
**Authentication Pattern**: Use cookie-based sessions (not JWT) - follows SSR best practices
**Styling**: Tailwind CSS
**Database**: PostgreSQL with GORM
**Framework**: Echo v4

## Development Principles

- Prefer server-side rendering over client-side solutions
- Follow traditional web app patterns (cookies, sessions, forms)
- Keep JavaScript minimal (HTMX or STIMULUS for dynamic behavior)
- Use existing Echo/GORM patterns established in codebase
- Follow existing code conventions and file structure

## Current State

The application has a working foundation with portal and QR code management, admin interface, and database migrations.

## Next Steps for Development

When source code is added, typical Go project commands will likely include:
- `go mod init` - Initialize Go module
- `go build` - Build the application  
- `go test ./...` - Run tests
- `go run main.go` - Run the application

Database connection will be available on the internal docker network to the PostgreSQL service.