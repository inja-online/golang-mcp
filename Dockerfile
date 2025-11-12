# syntax=docker/dockerfile:1.4
FROM golang:1.23-alpine AS builder

ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE

WORKDIR /build

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with cache mount for faster rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy source code
COPY . .

# Build with cache mounts for both module and build cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} \
    go build -a -installsuffix cgo \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT_SHA} -X main.date=${BUILD_DATE}" \
    -trimpath \
    -o mcp-go ./cmd/mcp-go

FROM alpine:latest

ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/mcp-go .

LABEL org.opencontainers.image.title="MCP Go Server" \
      org.opencontainers.image.description="Model Context Protocol server for Go development" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT_SHA}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/inja-online/golang-mcp"

ENTRYPOINT ["./mcp-go"]
