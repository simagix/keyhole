FROM golang:1.19-alpine as builder
RUN apk update && apk add git bash && rm -rf /var/cache/apk/* \  
  && mkdir -p /github.com/simagix/keyhole && cd /github.com/simagix \
  && git clone --depth 1 https://github.com/simagix/keyhole.git
WORKDIR /github.com/simagix/keyhole
RUN ./build.sh cross-platform
FROM alpine
LABEL Ken Chen <ken.chen@simagix.com>
RUN addgroup -S simagix && adduser -S simagix -G simagix
USER simagix
WORKDIR /dist
COPY --from=builder /github.com/simagix/keyhole/dist/keyhole-* /dist/
WORKDIR /home/simagix
COPY --from=builder /github.com/simagix/keyhole/dist/keyhole-linux-x64 /keyhole
CMD ["/keyhole", "--version"]
