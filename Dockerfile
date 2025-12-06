#打包过的文件直接copy到镜像里执行
FROM alpine

#COPY ./build/out/linux/* ./app
#EXPOSE 9000
#CMD ./gin_hello

#整个项目copy到镜像里编译执行
#FROM golang:latest
#WORKDIR /app
#COPY . /app
#RUN go build .
#EXPOSE 9000
#ENTRYPOINT ["./GinHello"]

#docker启动
#docker build -t gin-hello-alpine . && docker run -p 9000:9000 --name gin-hello-alpine gin-hello-alpine