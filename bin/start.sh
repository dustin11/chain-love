#!/bin/bash

PROJECT_NAME=$1
PROJECT_PATH=$2

#PID=$(ps aux | grep $JAR_NAME | grep -v 'grep' | awk '{print $2}')
#if [ ! -z "$PID" ];then
#    echo "$PROJECT_NAME 的进程ID是：$PID"
#        echo "即将关闭$PROJECT_NAME"
#    kill -9 $PID
#else
#    echo "$PROJECT_NAME 没有启动"
#fi
echo "即将关闭$PROJECT_NAME"
docker-compose down

echo "开始启动$PROJECT_NAME"
echo "path $PROJECT_PATH"
cd $PROJECT_PATH
docker-compose up --build

echo "项目发布完成"