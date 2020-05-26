FROM golang:alpine as builder
RUN apk update && apk add dep git && rm -rf /var/cache/apk/* \
  && mkdir -p /go/src/github.com/simagix/keyhole
ADD . /go/src/github.com/simagix/keyhole
WORKDIR /go/src/github.com/simagix/keyhole
RUN dep ensure && go build -o keyhole-linux-x64
FROM alpine
MAINTAINER Ken Chen <simagix@gmail.com>
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /home/simagix
COPY --from=builder /go/src/github.com/simagix/keyhole/keyhole-linux-x64 /
CMD ["/keyhole", "--version"]
