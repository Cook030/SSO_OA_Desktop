#!/bin/sh
set -e

echo "=== 等待 MySQL 就绪 ==="
/app/scripts/wait.sh

echo "=== 开始执行接口自动化测试 ==="
/app/test-binary -test.v -test.count=1
