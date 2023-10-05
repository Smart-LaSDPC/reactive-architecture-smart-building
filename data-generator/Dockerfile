# Parent image
FROM python:3

# Diretory that contains the program
WORKDIR /

# Copy the directory to the container 
COPY . .

# Install the MQTT
RUN pip3 install paho-mqtt

# Commands
#172.100.10.10
#159.100.10.10 Broket MQTT no Docker
CMD ["python3", "data-generator.py", "-h", "159.100.10.10", "-v", "-i", "02", "-n", "testclient1"]
