#!/bin/sh
set -eu

PROJECT_NAME=${1:-senspace}
PROJECT_PATH=${2:-$(pwd)}

echo "开始启动 $PROJECT_NAME"
echo "path $PROJECT_PATH"

if [ -d "$PROJECT_PATH/deploy" ]; then
  DEPLOY_DIR="$PROJECT_PATH/deploy"
elif [ -d "$PROJECT_PATH/../deploy" ]; then
  DEPLOY_DIR="$PROJECT_PATH/../deploy"
else
  echo "未找到 deploy 目录"
  exit 1
fi

cd "$DEPLOY_DIR"
docker compose down
docker compose --profile builder build plugin-builder
docker compose up -d --build mysql api nginx

echo "项目启动完成"
