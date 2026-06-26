# syntax=docker/dockerfile:1

FROM golang:1.25 AS builder

WORKDIR /build

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Transfer source code
COPY .git/refs/heads/main ./commitID
COPY *.go ./
COPY src ./src

# Build
ENV CGO_ENABLED=0
RUN commit_hash=$(cat commitID | cut -c1-7) && \
    go build -ldflags "-X github.com/birabittoh/disgord/src/globals.CommitID=$commit_hash" -o /dist/disgord

# Install playwright firefox for ARL auto-renewal
RUN go run github.com/playwright-community/playwright-go/cmd/playwright install firefox

# Test
FROM builder AS run-test-stage
RUN go test -v ./...

FROM debian:bookworm-slim AS build-release-stage

# Firefox runtime deps + ffmpeg
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    ffmpeg \
    libgtk-3-0 libdbus-glib-1-2 libxt6 libasound2 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY templates ./templates
COPY --from=builder /dist .
COPY --from=builder /root/.cache/ms-playwright /root/.cache/ms-playwright
COPY --from=builder /root/.cache/ms-playwright-go /root/.cache/ms-playwright-go

ENTRYPOINT ["./disgord"]
