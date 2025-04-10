all:
	GOPROXY=https://goproxy.cn,direct CGO_ENABLE=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build '-buildmode=pie' -ldflags '-extldflags -static' -gcflags ''  -o ./sayhi

clean:
	rm -rf ./sayhi

docker:
	docker build -t registry.cn-hangzhou.aliyuncs.com/xsw1058/sayhi:latest .

