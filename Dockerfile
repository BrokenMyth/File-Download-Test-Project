# 使用 golang 官方镜像作为基础镜像
FROM golang:latest

# 设置工作目录
WORKDIR /go/src/app

# 将应用程序添加到工作目录
ADD . /go/src/app

# 编译应用程序
RUN go build -o myapp .

# 暴露端口
EXPOSE 8080

# 启动应用程序
CMD ["./myapp"]