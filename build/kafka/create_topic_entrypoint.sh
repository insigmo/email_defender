#!/bin/bash
set -e

/opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server kafka1:9092 \
  --create \
  --if-not-exists \
  --topic "$KAFKA_TOPIC" \
  --partitions 3 \
  --replication-factor 3

echo 'Topics created'