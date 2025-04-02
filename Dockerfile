# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder


WORKDIR /build

# Download Git submodules
# COPY .git ./.git
# RUN git -c submodule.ui.update=none submodule update --init --recursive

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Transfer source code
COPY .git/refs/heads/main ./commitID
COPY *.go ./
COPY src ./src

# Build
RUN commit_hash=$(cat commitID | cut -c1-7) && \
    CGO_ENABLED=0 go build -ldflags "-X github.com/birabittoh/disgord/src/globals.CommitID=$commit_hash" -trimpath -o /dist/disgord


# Test
FROM builder AS run-test-stage
# COPY i18n ./i18n
RUN go test -v ./...

FROM alpine AS build-release-stage

RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY --from=builder /dist .
ENTRYPOINT ["./disgord"]
