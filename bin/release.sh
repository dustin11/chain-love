#!/bin/bash

PROJECT_NAME="smarterp"
PROJECT_PATH="/usr/local/app/$PROJECT_NAME"
APP_PATH="bin/out/linux/$PROJECT_NAME"
SHELL_NAME="start.sh"
LOG_PATH="/var/log/$PROJECT_NAME"

BASEDIR=$1
SERVICE=$2
ENV=$3

echo "开始部署$ENV 环境"
cd $BASEDIR

echo "从git更新代码"
#git pull

echo "开始编译代码 mac - linux"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $APP_PATH main.go

echo "开始复制$PROJECT_NAME 到服务器$SERVICE"
ssh root@$SERVICE "mkdir -p $PROJECT_PATH $LOG_PATH"
scp -r $APP_PATH mysql Dockerfile docker-compose.yml bin/$SHELL_NAME root@$SERVICE:$PROJECT_PATH

ssh root@$SERVICE "cd $PROJECT_PATH && sh $SHELL_NAME $PROJECT_NAME $PROJECT_PATH $ENV"