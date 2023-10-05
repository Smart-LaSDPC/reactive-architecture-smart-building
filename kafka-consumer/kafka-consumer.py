from kafka import KafkaConsumer
import json
import avro
from avro_json_serializer import AvroJsonSerializer
import msgpack

schema_dict = {
    "namespace": "signals.avro",
        "type": "record",
        "name": "Signal",
        "fields": [
            {"name":"date", "type":"string"},
            {"name":"agent_id", "type":"string"},
            {"name":"temperature", "type": ["long", "int", "double"]},
            {"name":"moisture", "type":["long", "int", "double"]},
            {"name":"state", "type":"string"}
        ]
}

avro_schema = avro.schema.make_avsc_object(schema_dict, avro.schema.Names())
serializer = AvroJsonSerializer(avro_schema)

# To consume latest messages and auto-commit offsets
consumer = KafkaConsumer('signals',
                         group_id='my-group',
                         bootstrap_servers=['159.100.10.5:9092'],
                         reconnect_backoff_ms=50)
for message in consumer:
    # message value and key are raw bytes -- decode if necessary!
    # e.g., for unicode: `message.value.decode('utf-8')`
    # print ("%s:%d:%d: key=%s value=%s" % (message.topic, message.partition,
    #                                      message.offset, message.key,
    #                                      message.value.decode('utf-8')))
    print(message.value.decode('utf-8'))

# consume earliest available messages, don't commit offsets
KafkaConsumer(auto_offset_reset='earliest', enable_auto_commit=False)

# consume json messages
# KafkaConsumer(value_deserializer=lambda m: json.loads(m.decode('ascii')))
KafkaConsumer(value_deserializer=lambda m: msgpack.unpackb(serializer.to_json(m)))

# consume msgpack
# KafkaConsumer(value_deserializer=msgpack.unpackb)

# StopIteration if no message after 1sec
KafkaConsumer(consumer_timeout_ms=1000)

# Subscribe to a regex topic pattern
consumer = KafkaConsumer()
consumer.subscribe(pattern='^awesome.*')

# Use multiple consumers in parallel w/ 0.9 kafka brokers
# typically you would run each on a different server / process / CPU
# consumer1 = KafkaConsumer('my-topic', group_id='my-group', bootstrap_servers='my.server.com')
# consumer2 = KafkaConsumer('my-topic', group_id='my-group', bootstrap_servers='my.server.com')