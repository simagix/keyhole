FROM golang:1.13-alpine as builder
RUN apk update && apk add dep git bash && rm -rf /var/cache/apk/* \
  && mkdir -p /go/src/github.com/simagix/keyhole
ADD . /go/src/github.com/simagix/keyhole
WORKDIR /go/src/github.com/simagix/keyhole
RUN ./build.sh all
FROM alpine
MAINTAINER Ken Chen <simagix@gmail.com>
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /build
COPY --from=builder /go/src/github.com/simagix/keyhole/build/keyhole-* /build/
WORKDIR /home/simagix
COPY --from=builder /go/src/github.com/simagix/keyhole/build/keyhole-linux-x64 /keyhole
CMD ["/keyhole", "--version"]
