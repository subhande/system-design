## Remove broker and kafka-ui containers

docker stop broker
docker rm broker

docker stop kafka-ui
docker rm kafka-ui

docker run -d -it --name broker -p 9092:9092 -p 9094:9094 \
-e KAFKA_NODE_ID=1 \
-e KAFKA_PROCESS_ROLES=broker,controller \
-e KAFKA_LISTENERS=PLAINTEXT://:9092,PLAINTEXT_HOST://:9094,CONTROLLER://:9093 \
-e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://host.docker.internal:9092,PLAINTEXT_HOST://localhost:9094 \
-e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
-e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT \
-e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
-e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
-e KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 \
-e KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
-e KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0 \
-e KAFKA_NUM_PARTITIONS=3 \
apache/kafka:latest


docker run -d -it --name kafka-ui -p 8080:8080 \
-e KAFKA_BROKERS=host.docker.internal:9092 \
docker.redpanda.com/redpandadata/console:latest