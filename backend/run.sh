#!/bin/sh
set -e

# 后端服务启动脚本

cd /app

# 生产环境建议通过挂载的方式覆盖 /app/config/config.yaml
# docker run -v /path/to/config.yaml:/app/config/config.yaml ...

exec ./main
