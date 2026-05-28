# ── Stage 1: build Svelte frontend ────────────────────────────────────────────
FROM node:22-alpine AS frontend

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm ci --prefer-offline

COPY web/ ./
RUN npm run build


# ── Stage 2: build Go binary ───────────────────────────────────────────────────
FROM golang:1.25-alpine AS backend

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -extldflags=-static" -o /app/machino ./cmd/api


# ── Stage 3: minimal runtime image ────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S machino \
    && adduser  -S -G machino machino

WORKDIR /app

COPY --from=backend /app/machino ./machino
COPY --from=frontend /app/web/dist ./web/dist

# SQLite data lives in a mounted volume
RUN mkdir /data && chown machino:machino /data
VOLUME ["/data"]

ENV DATABASE_PATH=/data/machino.db
ENV STATIC_DIR=/app/web/dist
ENV HTTP_ADDR=:8080
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json

USER machino

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/health || exit 1

ENTRYPOINT ["/app/machino"]
