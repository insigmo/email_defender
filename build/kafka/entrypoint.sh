#!/bin/bash
set -e

/opt/kafka/bin/kafka-storage.sh format \
  -t THE_KAFKA_RAFT_CLUSTER \
  -c /var/lib/kafka/kraft.properties \
  --ignore-formatted

echo "==> Starting Kafka server..."
exec /opt/kafka/bin/kafka-server-start.sh /var/lib/kafka/kraft.properties