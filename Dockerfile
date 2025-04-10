FROM registry.cn-hangzhou.aliyuncs.com/xsw1058/golang:1.23-alpine AS buildxusw
ADD ./* ./
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk update && \
    apk add gpgme btrfs-progs-dev llvm15-dev gcc musl-dev
RUN GOPROXY=https://goproxy.cn,direct CGO_ENABLE=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build '-buildmode=pie' -ldflags '-extldflags -static' -gcflags ''  -o ./sayhi

FROM registry.cn-hangzhou.aliyuncs.com/xsw1058/alpine:3.21.3
COPY --from=buildxusw /go/sayhi      /usr/local/bin/sayhi
ENTRYPOINT ["/usr/local/bin/sayhi"]