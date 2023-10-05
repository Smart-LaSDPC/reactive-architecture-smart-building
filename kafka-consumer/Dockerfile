# Parent image
FROM python:3

# Diretory that contains the program
WORKDIR /

# Copy the directory to the container 
COPY . .

# Install the MQTT
RUN pip3 install msgpack kafka-python avro-json-serializer avro

# Commands
CMD ["python3", "kafka-consumer.py"]
