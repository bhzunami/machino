# ── Stage 1: build Svelte frontend ────────────────────────────────────────────
FROM node:22-alpine AS frontend

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build


# ── Stage 2: build Go binary ───────────────────────────────────────────────────
FROM golang:1.25-alpine AS backend

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/machino ./cmd/api


# ── Stage 3: minimal runtime image ────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend /app/machino ./machino
COPY --from=frontend /app/web/dist ./web/dist

# SQLite data lives in a mounted volume
VOLUME ["/data"]

ENV DATABASE_PATH=/data/machino.db
ENV STATIC_DIR=/app/web/dist
ENV HTTP_ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["/app/machino"]
