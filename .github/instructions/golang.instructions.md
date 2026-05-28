---
description: These instructions should be followed by the AI when generating GO code, answering questions, or reviewing changes for this project.
applyTo: '**/*.go'
# applyTo: 'Describe when these instructions should be loaded by the agent based on task context' # when provided, instructions will automatically be added to the request context when the pattern matches an attached file
---

<!-- Tip: Use /create-instructions in chat to generate content with agent assistance -->
# Copilot Instructions

## Stack
Go 1.25, gorilla/mux, gorilla/websocket, pgx (PostgreSQL), sqlc for queries, slog for logging.

## Code Style
- Follow Effective Go and Go Code Review Comments
- Use table-driven tests
- Error wrapping: fmt.Errorf("functionName: %w", err)
- No naked returns

## Project Structure
- cmd/ - Entry points (cmd/api/main.go)
- internal/ - Private application code
- internal/handler/ - HTTP handlers
- internal/service/ - Business logic
- internal/repository/ - Database layer (sqlc generated)
- internal/middleware/ - HTTP middleware

## Patterns
- Handlers accept (w http.ResponseWriter, r *http.Request)
- Use context.Context for cancellation and request-scoped values
- Dependency injection via struct fields, not globals
- JSON responses with a respond() helper, not manual encoding

## Avoid
- Do not use global variables or init() functions
- Do not use panic for error handling
- Do not use external ORM libraries -- use sqlc