# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Timely** is a Go-based calendar management backend API with Google Calendar integration, featuring JWT authentication and clean architecture patterns. The project uses Chi router, GORM ORM with SQLite, and comprehensive OAuth2 implementation.

## Development Commands

### Backend (Go)
```bash
# Development - starts server on port 8000
go run main.go

# Testing
go test ./...                    # Run all tests
go test ./pkg/...               # Run package tests only

# Build
go build -o timely main.go      # Build production binary

# Dependencies
go mod tidy                     # Clean up dependencies
go mod download                 # Download dependencies

# Documentation
# Swagger docs available at localhost:8000/swagger/ in development mode
```

## Architecture Overview

### Clean Architecture Pattern
- **Handlers** (`internal/handler/`): HTTP request handling and response formatting
- **Services** (`internal/service/`): Business logic and external API integration
- **Repositories** (`internal/repository/`): Data access and database operations
- **Models** (`internal/model/`): Data structures and DTOs

### Key Components

**Authentication System:**
- Traditional email/password auth with Argon2 hashing
- Google OAuth2 integration with secure state management
- JWT tokens for stateless authentication
- Account linking for multiple auth providers

**Calendar Integration:**
- Google Calendar API integration
- Calendar import and event synchronization
- Multi-calendar support with visibility controls

**Database Schema:**
- `users` - User accounts with profile information
- `accounts` - OAuth account linking (Google, etc.)
- `calendars` - Calendar metadata and source tracking
- `calendar_events` - Individual calendar events with timing

### Core Dependencies
- **Router:** `github.com/go-chi/chi/v5` - HTTP routing and middleware
- **ORM:** `gorm.io/gorm` with SQLite driver
- **Auth:** `github.com/golang-jwt/jwt/v5`, `golang.org/x/oauth2`
- **Logging:** `go.uber.org/zap` with log rotation
- **Documentation:** `github.com/swaggo/swag` for Swagger generation
- **IDs:** `github.com/bwmarrin/snowflake` for distributed unique IDs
- **ICS Processing:** `github.com/arran4/golang-ical` for ICS/iCal file parsing

## API Endpoints

### Authentication (`/api/auth`)
- `POST /api/auth/login` - Traditional login
- `POST /api/auth/register` - User registration  
- `GET /api/auth/google/login` - Google OAuth initiation
- `GET /api/auth/google/callback` - Google OAuth callback

### Calendar (`/api/calendar`) - JWT Protected
- `GET /api/calendar/google/` - Get Google calendars
- `POST /api/calendar/google/` - Import Google calendar
- `POST /api/calendar/ics/import` - Import ICS file via JSON body
- `POST /api/calendar/ics/upload` - Import ICS file via multipart form upload
- `GET /api/calendar/events` - Get calendar events with time range filtering

### System
- `GET /health` - Health check endpoint
- `GET /swagger/*` - API documentation (development only)

## Configuration

Required environment variables in `.env`:
- `GO_ENV` - Environment mode (development/production)
- `JWT_SECRET` - JWT signing secret
- `GOOGLE_CLIENT_ID` - Google OAuth client ID
- `GOOGLE_CLIENT_SECRET` - Google OAuth client secret
- `GOOGLE_REDIRECT_URL` - Google OAuth callback URL
- `FRONTEND_DOMAIN` - Frontend domain for CORS configuration

## Testing

Tests are located in `pkg/` subdirectories following Go conventions:
- Password hashing validation (`pkg/encrypt/`)
- OAuth state management (`pkg/oauth/`)
- ID generation (`pkg/utils/`)
- Logging functionality (`pkg/logger/`)

## Database

- **Database:** SQLite with GORM auto-migration
- **File:** `timely.db` (created automatically)
- **Migration:** Automatic on startup via `DatabaseInit()` in `cmd/app.go:107`

## Development Notes

- Server runs on port 8000 by default
- Swagger documentation available in development mode at `/swagger/`
- CORS configured for frontend integration
- Snowflake ID generation initialized with node ID 1
- Structured logging with Zap logger and file rotation
- Clean separation of authentication and calendar functionality