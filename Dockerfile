# Copyright 2020 Kuei-chun Chen. All rights reserved.
FROM golang:1.25-alpine AS builder

# Install dependencies first (cached layer)
RUN apk update && apk add git bash && rm -rf /var/cache/apk/*

WORKDIR /github.com/simagix/keyhole

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build binary (OS and arch auto-detected by buildx platform)
RUN ./build.sh binary

FROM alpine:3.19
LABEL maintainer="Ken Chen <ken.chen@simagix.com>"
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /home/simagix
COPY --from=builder /github.com/simagix/keyhole/keyhole /keyhole

CMD ["/keyhole", "--version"]
