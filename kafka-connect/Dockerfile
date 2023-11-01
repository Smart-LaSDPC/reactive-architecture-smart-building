FROM confluentinc/cp-kafka-connect:7.1.0-1-ubi8

ENV CONNECT_PLUGIN_PATH: "/usr/share/java,/usr/share/confluent-hub-components"

RUN confluent-hub install --no-prompt confluentinc/kafka-connect-mqtt:latest
RUN confluent-hub install --no-prompt debezium/debezium-connector-postgresql:2.2.1
