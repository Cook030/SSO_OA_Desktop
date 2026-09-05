#!/bin/sh
# 等待 MySQL 就绪（最多 60 秒）
set -e

HOST="${MYSQL_HOST:-mysql}"
PORT="${MYSQL_PORT:-3306}"
USER="${MYSQL_USER:-root}"
PASS="${MYSQL_PASS:-test_password_123}"
DB="${MYSQL_DB:-maplehaze_permission_test}"

MAX_RETRIES=30
RETRY_INTERVAL=2

echo "等待 MySQL ($HOST:$PORT) 就绪..."

for i in $(seq 1 $MAX_RETRIES); do
    if mysqladmin ping -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" --silent 2>/dev/null; then
        echo "MySQL 已就绪 (尝试 $i/$MAX_RETRIES)"
        exit 0
    fi
    echo "  尝试 $i/$MAX_RETRIES - MySQL 未就绪，${RETRY_INTERVAL}s 后重试..."
    sleep $RETRY_INTERVAL
done

echo "错误: MySQL 在 $((MAX_RETRIES * RETRY_INTERVAL))s 内未就绪" >&2
exit 1
