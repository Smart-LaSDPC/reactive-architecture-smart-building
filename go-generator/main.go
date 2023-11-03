package main

import (
	"fmt"
	"sync"
	"time"
	"strconv"
	"context"
	"os"
	"os/signal"
	"syscall"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	brokerAddr 	  := ReadEnvVariable("GENERATOR_BROKER_ADDRESS", "localhost:1883")
	topic	   	  := ReadEnvVariable("GENERATOR_TOPIC", "farm_id/device_id/testclient1")
	
	numDevicesStr := ReadEnvVariable("GENERATOR_NUM_DEVICES", "10")
	numDevices, err := strconv.Atoi(numDevicesStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_NUM_DEVICES env variable: %s", err))
	}

	publishTimeoutMsStr := ReadEnvVariable("GENERATOR_PUBLISH_TIMEOUT_MS", "10")
	publishTimeoutMs, err := strconv.Atoi(publishTimeoutMsStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_PUBLISH_TIMEOUT_MS env variable: %s", err))
	}

	continuousStr := ReadEnvVariable("GENERATOR_PUBLISH_CONTINUOUS", "true")
	continuous, err := strconv.ParseBool(continuousStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_PUBLISH_CONTINUOUS env variable: %s", err))
	}

	durationStr := ReadEnvVariable("GENERATOR_PUBLISH_DURATION_SECONDS", "10")
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_PUBLISH_DURATION_SECONDS env variable: %s", err))
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	start := make(chan interface{})
	end := make(chan interface{})

	messagesGenerated := 0

	for i := 0; i < numDevices; i++ {
		deviceID := fmt.Sprintf("%d", i)
		mqttOptions := mqtt.NewClientOptions()
		mqttOptions.AddBroker(brokerAddr)
		publishTimeout := time.Duration(publishTimeoutMs) * time.Millisecond

		device, err := NewDevice(deviceID, topic, publishTimeout, mqttOptions)
		if err != nil {
			fmt.Printf("Failed to connect device %s: %s\n", deviceID, err)
			continue
		}
		fmt.Printf("Succesfully connected Device %s\n", deviceID)

		wg.Add(1)
		go device.PublishData(ctx, start, wg, &messagesGenerated)
	}

	// Sinal para comecar a publicar mensagens
	close(start)

	// Gorotina para terminar a execucao quando o tempo terminar
	if !continuous {
		go func(){
			time.Sleep(time.Duration(duration) * time.Second)
			close(end)
		}()
	}

	keepRunning := true
	for keepRunning {
		select {
		case <-sigterm:
			fmt.Println("Terminating: via signal")
			keepRunning = false
		case <-end:
			fmt.Printf("Terminating: completed after %d seconds\n", duration)
			keepRunning = false
		}
	}
	cancel()
	wg.Wait()

	fmt.Println("### End of MQTT synthetic generation")
	fmt.Printf("### Generated %d messages\n", messagesGenerated)
}

func ReadEnvVariable(env, fallback string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		return fallback
	}
	return v
}
