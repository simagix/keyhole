FROM golang:1.25-alpine AS builder
RUN apk update && apk add git bash && rm -rf /var/cache/apk/*
WORKDIR /github.com/simagix/keyhole

# Copy go.mod first for dependency caching
COPY go.mod ./
RUN go mod download

# Copy source files
COPY . .
RUN ./build.sh cross-platform
FROM alpine
LABEL maintainer="Ken Chen <ken.chen@simagix.com>"
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /dist
COPY --from=builder /github.com/simagix/keyhole/dist/keyhole-* /dist/
WORKDIR /home/simagix
COPY --from=builder /github.com/simagix/keyhole/dist/keyhole-linux-x64 /keyhole
CMD ["/keyhole", "--version"]
