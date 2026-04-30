#!/bin/bash
set -e


/opt/kafka/bin/kafka-storage.sh format \
  -t THE_KAFKA_RAFT_CLUSTER \
  -c /var/lib/kafka/kraft.properties \
  --ignore-formatted

#/opt/kafka/bin/kafka-topics.sh --create --topic parted --bootstrap-server localhost:19092,localhost:29092,localhost:39092

echo "==> Starting Kafka server..."
/opt/kafka/bin/kafka-server-start.sh /var/lib/kafka/kraft.properties
