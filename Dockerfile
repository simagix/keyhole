FROM alpine
MAINTAINER Ken Chen <simagix@gmail.com>
ADD build/keyhole-linux-x64 /keyhole
CMD ["/keyhole", "--version"]
