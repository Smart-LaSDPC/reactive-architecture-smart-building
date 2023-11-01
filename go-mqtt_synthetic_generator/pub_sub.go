package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Message %s received on topic %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}

func sub_generate(deviceID, broker, topic string, iterations int) {
	clientID := time.Now().String() + " DeviceID: " + deviceID
	// MQTT Client Options
	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID(clientID)
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println(clientID + "connected")

	token = client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic %s\n", topic)

	for i := 0; i < iterations; i++ {
		text := fmt.Sprintf("%d", i)
		token = client.Publish(topic, 0, false, text)
		token.Wait()
		time.Sleep(10 * time.Millisecond)
	}

	defer client.Disconnect(100)
}
