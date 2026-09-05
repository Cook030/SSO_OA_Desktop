#!/usr/bin/env bash
# 创建审计 binlog topic(幂等)。若启用了 auto-create(见 docker-compose KAFKA_AUTO_CREATE_TOPICS_ENABLE),
# 首次生产消息会自动创建, 但分区数取 broker 默认值; 建议显式创建以保证分区数与 canal.mq.partitionsNum 一致。
set -euo pipefail

BOOTSTRAP="${BOOTSTRAP_SERVER:-localhost:9092}"
TOPIC="${TOPIC:-oasso.audit.binlog}"
PARTITIONS="${PARTITIONS:-3}"
REPLICATION_FACTOR="${REPLICATION_FACTOR:-1}"

docker exec -it audit-kafka kafka-topics \
  --create \
  --bootstrap-server "$BOOTSTRAP" \
  --topic "$TOPIC" \
  --partitions "$PARTITIONS" \
  --replication-factor "$REPLICATION_FACTOR" \
  --if-not-exists
