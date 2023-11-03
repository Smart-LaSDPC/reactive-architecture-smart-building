package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	brokerAddr := ReadEnvVariable("GENERATOR_BROKER_ADDRESS", "localhost:1883")
	topic := ReadEnvVariable("GENERATOR_TOPIC", "farm_id/device_id/testclient1")

	numDevicesStr := ReadEnvVariable("GENERATOR_NUM_DEVICES", "10")
	numDevices, err := strconv.Atoi(numDevicesStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_NUM_DEVICES env variable: %s", err))
	}

	msgsPerSecondStr := ReadEnvVariable("GENERATOR_MSGS_PER_SECOND", "10")
	msgsPerSecond, err := strconv.Atoi(msgsPerSecondStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_MSGS_PER_SECOND env variable: %s", err))
	}

	totalMsgsStr := ReadEnvVariable("GENERATOR_TOTAL_MESSAGES", "10")
	totalMsgs, err := strconv.Atoi(totalMsgsStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_TOTAL_MESSAGES env variable: %s", err))
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	start := make(chan interface{})
	end := make(chan interface{})

	var startTime time.Time
	var elapsedTime time.Duration

	messagesGenerated := 0

	fmt.Printf("### Start of MQTT synthetic generation\n")
	fmt.Printf("### Simulating %d devices publishing at a rate of %d messages/second until %d messages are published\n", numDevices, msgsPerSecond, totalMsgs)

	for i := 0; i < numDevices; i++ {
		deviceID := fmt.Sprintf("%d", i)
		mqttOptions := mqtt.NewClientOptions()
		mqttOptions.AddBroker(brokerAddr)
		publishTimeoutMs := (1000 / time.Duration(msgsPerSecond)) * time.Millisecond

		device, err := NewDevice(deviceID, topic, publishTimeoutMs, mqttOptions)
		if err != nil {
			fmt.Printf("Failed to connect device %s: %s\n", deviceID, err)
			continue
		}
		fmt.Printf("Succesfully connected Device %s\n", deviceID)

		wg.Add(1)
		go device.PublishData(ctx, start, wg, &messagesGenerated, &totalMsgs)
	}

	fmt.Printf("### All devices connected. Starting to publish messages.\n")

	// Sinal para comecar a publicar mensagens
	close(start)
	startTime = time.Now()

	// Gorotina para terminar a execucao quando chegar ao maximo de mensagens
	if totalMsgs > 0 {
		go func() {
			for {
				if messagesGenerated >= totalMsgs {
					close(end)
					elapsedTime = time.Since(startTime)
					return
				}
			}
		}()
	}

	keepRunning := true
	for keepRunning {
		select {
		case <-sigterm:
			fmt.Println("Terminating: via signal")
			keepRunning = false
		case <-end:
			fmt.Printf("Terminating: completed\n")
			keepRunning = false
		}
	}
	cancel()
	wg.Wait()

	fmt.Printf("### End of MQTT synthetic generation\n")
	fmt.Printf("### Generated %d messages in %s\n", messagesGenerated, elapsedTime)
}

func ReadEnvVariable(env, fallback string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		return fallback
	}
	return v
}
