package main

import (
	"fmt"
	"sync"
	"log"
	"time"
	"context"
	"math/rand"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection Lost: %s\n", err.Error())
}

type Device struct {
	DeviceID       string
	Topic          string
	publishTimeout time.Duration
	mqttClient     mqtt.Client
}

func NewDevice(deviceID string, topic string, publishTimeout time.Duration, mqttOptions *mqtt.ClientOptions) (*Device, error) {
	clientID := fmt.Sprintf("%s DeviceID: %s", time.Now().String(), deviceID)

	mqttOptions.SetClientID(clientID)
	mqttOptions.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(mqttOptions)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("Failed to connect mqtt client: %s", token.Error())
	}

	return &Device{
		DeviceID:       deviceID,
		Topic:          topic,
		publishTimeout: publishTimeout,
		mqttClient:     client,
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

func (d *Device) PublishData(ctx context.Context, start <-chan interface{}, wg *sync.WaitGroup, messagesGenerated, totalMsgs *int) {
	defer wg.Done()
	defer d.mqttClient.Disconnect(100)

	// Bloqueando ate o sinal de inicio
	<-start

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if *messagesGenerated < *totalMsgs || *totalMsgs <= 0 {
				*messagesGenerated += 1
				payload := d.generateMsgPayload()
				token := d.mqttClient.Publish(d.Topic, 0, false, payload)
				token.Wait()
				time.Sleep(d.publishTimeout)
			}
		}
	}
}
