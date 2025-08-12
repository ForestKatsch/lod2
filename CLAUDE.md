# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

lod2 is a Go web application that powers [lod2.zip](https://lod2.zip/). It's built with the chi router and uses SQLite for data storage. The application focuses on graceful degradation and self-sufficiency with minimal external dependencies.

## Development Commands

### Building and Running
- `go build` - Compile the application
- `go build -ldflags "-X 'lod2/page.BuildTime=$(date +%Y%m%d%H%M%S)'"` - Compile with cache-busting version
- `go run main.go` - Run the application directly (listens on localhost:10800 by default)
- `go run main.go -host 0.0.0.0 -port 8080` - Run with custom host/port
- `go mod tidy` - Clean up module dependencies

### Testing and Quality
- `go test ./...` - Run all tests
- `go vet ./...` - Run static analysis
- `go fmt ./...` - Format code

### Configuration
- `-config` flag: Configuration directory (default: `~/.config/lod2/`)
- `-data` flag: Data directory (default: `~/.local/share/lod2/`)

## Architecture

### Core Structure
- **main.go**: Application entry point, sets up chi router with middleware and host-based routing
- **config/**: Configuration management with CLI flags
- **db/**: SQLite database initialization and migrations
- **auth/**: JWT-based authentication system with sessions, users, roles, and access tokens
- **routes/**: HTTP route handlers organized by domain (auth, account, admin)
- **page/**: Template rendering system using Go's html/template with Sprig functions
- **middleware/**: Custom middleware (auth refresh)
- **cplane/**: Control plane for webhooks and redeployment (separate subdomain routing)
- **templates/**: HTML templates with library/layout system and individual pages
- **static/**: Static assets (CSS, JS, fonts, images)

### Key Patterns

#### Host-based Routing
The application uses hostrouter to serve different content based on subdomain:
- `cplane.lod2.zip` → Control plane routes (webhooks, status, redeploy)
- `*` (all other hosts) → Main application routes

#### Template System
- Templates are organized in `templates/library/` (shared components) and `templates/pages/` (page-specific)
- Page rendering uses a metadata system that injects user info, timestamps, and request data
- Template caching can be disabled via `enablePageCache` constant in page/page.go:21

#### Authentication Architecture
- JWT-based with RS256 signing using JWK keys
- Session management with database storage
- User roles and access control
- Middleware for automatic token refresh
- Private key stored at `~/.config/lod2/keys/auth/private.jwk.json`

#### Database
- SQLite with automatic migrations
- Database file stored at `~/.local/share/lod2/lod2.db`
- Migration system handled in auth and db packages

### Frontend
- HTMX for dynamic interactions (htmx.min.js)
- Custom CSS with paper-style design system
- Environment banner shows "LOCAL" when running on localhost
- Dark mode support implemented in stylesheets

## Required Setup Files

The application requires a private JWK key for JWT signing:
- Path: `~/.config/lod2/keys/auth/private.jwk.json`
- Generate at [mkjwk.org](https://mkjwk.org/) with Key Use: Signature, Algorithm: RS256
- Use the "Public and Private Keypair" JSON object as file contents