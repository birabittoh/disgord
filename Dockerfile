# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

# RUN apk add --no-cache git

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
    CGO_ENABLED=0 go build -ldflags "-X github.com/BiRabittoh/disgord/src.CommitID=$commit_hash" -trimpath -o /dist/app


# Test
FROM builder AS run-test-stage
# COPY i18n ./i18n
RUN go test -v ./...

FROM scratch AS build-release-stage

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dist .
# COPY i18n ./i18n
# COPY publi[c] ./public

ENTRYPOINT ["./app"]
