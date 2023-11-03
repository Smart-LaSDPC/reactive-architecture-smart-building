package main

import (
	"fmt"
	"math/rand"
	"time"
	"context"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Message %s published on topic %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}

type Device struct {
	DeviceID string
	Topic string
	PublishTimeout time.Duration
	mqttClient mqtt.Client
}

func NewDevice(deviceID string, topic string, publishTimeout time.Duration, mqttOptions *mqtt.ClientOptions) (*Device, error) {
	clientID := fmt.Sprintf("%s DeviceID: %s", time.Now().String(), deviceID)
		
	mqttOptions.SetClientID(clientID)
	mqttOptions.SetDefaultPublishHandler(messagePubHandler)
	mqttOptions.OnConnect = connectHandler
	mqttOptions.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(mqttOptions)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("Failed to connect mqtt client: %s", token.Error())
	}

	// token = client.Subscribe(topic, 1, nil)
	// token.Wait()
	// fmt.Printf("Subscribed to topic %s\n", topic)

	return &Device{
		DeviceID: deviceID,
		Topic: topic,
		PublishTimeout: publishTimeout,
		mqttClient: client,
	}, nil	
}

func (d *Device) generateMsgPayload() string {
	date := time.Now().Format("01/02/2006 15:04:05")
	agentId := fmt.Sprintf("AGENT%s", d.DeviceID)
	temperature := rand.Intn(60)
	moisture := rand.Intn(60)
	state := "ON"

	return fmt.Sprintf(`{ "date":"%s","agent_id":"%s","temperature":%d,"moisture":%d,"state":"%s" }`, date, agentId, temperature, moisture, state)
}

func (d *Device) PublishData(ctx context.Context, start <-chan interface{}, wg *sync.WaitGroup, messagesGenerated *int) {
	defer wg.Done()
	defer d.mqttClient.Disconnect(100)

	// Block until start signal
	<-start
	
	fmt.Printf("Started publishing messages for device %s\n", d.DeviceID)
	for {
		select{
		case <-ctx.Done():
			fmt.Printf("Done publishing messages for device %s\n", d.DeviceID)
			return
		default:
			payload := d.generateMsgPayload()
			token := d.mqttClient.Publish(d.Topic, 0, false, payload)
			token.Wait()
			*messagesGenerated += 1
			time.Sleep(d.PublishTimeout)
		}
	}
}
