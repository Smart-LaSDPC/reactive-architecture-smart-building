##Display all sensor that is connected to Mosquito Broker
import ssl
import sys

import paho.mqtt.client as mqtt

def on_connect(client, userdata, flags, rc):
	print('connected (%s)' % client._client_id)
	client.subscribe(topic='#', qos=1)
	
def on_message(client, userdata, message):
	print('-----------------------------')
	print('topic: %s' % message.topic)
	print('payload: %s' % message.payload)
	print('qos: %d' % message.qos)

def main():
	#Create client object
	client = mqtt.Client(client_id='SUBSSS', clean_session=False)
	client.on_connect = on_connect
	client.on_message = on_message
	client.connect(host='159.100.10.10', port=1883)
	# client.connect(host='172.100.10.10', port=1883)
	# client.connect(host='127.0.0.0', port=1883)
	client.loop_forever()
	
if __name__ == '__main__':
	main()
	
sys.exit(0)		
