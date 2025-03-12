FROM golang:1.24.1-alpine3.20 AS build

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
COPY . /go/src/gocode
WORKDIR /go/src/gocode
RUN export GOPROXY=https://goproxy.cn && \
    go mod vendor && \
    go build -mod vendor


FROM alpine:3.20

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && apk update && apk add tzdata
WORKDIR /server
COPY --from=build /go/src/gocode/tcp-proxy /server/tcp-proxy
