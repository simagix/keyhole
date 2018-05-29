# Docker
Build docker image for *keyhole*.

## Docker Build

```
$ docker build -t simagix/keyhole .

REPOSITORY                         TAG                 IMAGE ID            CREATED             SIZE
simagix/keyhole                    latest              1fc381cbafd7        14 minutes ago      9.75MB

$ docker push simagix/keyhole
```

Image `simagix/keyhole` is also available from [Docker hub](https://hub.docker.com/).

### Lightweight Dockerfile
The image file is less than 10MB.

```
FROM alpine
MAINTAINER Ken Chen <simagix@gmail.com>
ADD build/keyhole-linux-x64 /keyhole
CMD ["/keyhole", "--version"]
```

## Docker Commands
### Check Version

```
$ docker run simagix/keyhole
keyhole ver. master-20180528.1527529455
```

### Get Info
Connect to an instance on the Docker host.

```
docker run -v /etc/ssl/certs:/etc/ssl/certs simagix/keyhole \
    /keyhole --uri mongodb://$(hostname -f):30000/ --info
```

### Atlas Example
```
docker run -v /etc/ssl/certs:/etc/ssl/certs simagix/keyhole \
    /keyhole --uri mongodb://admin:secret@cluster0-shard-00-01-nhftn.mongodb.net.:27017,cluster0-shard-00-02-nhftn.mongodb.net.:27017,cluster0-shard-00-00-nhftn.mongodb.net.:27017/test?replicaSet=Cluster0-shard-0\&authSource=admin \
    --ssl --info
```