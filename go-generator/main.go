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
	brokerAddr 			:= ReadEnvVariable("GENERATOR_BROKER_ADDRESS", "localhost:1883")
	topic		 		:= ReadEnvVariable("GENERATOR_TOPIC", "farm_id/device_id/testclient1")
	
	numDevicesStr 		:= ReadEnvVariable("GENERATOR_NUM_DEVICES", "10")
	numDevices, err := strconv.Atoi(numDevicesStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_NUM_DEVICES env variable: %s", err))
	}

	publishTimeoutMsStr := ReadEnvVariable("GENERATOR_PUBLISH_TIMEOUT_MS", "10")
	publishTimeoutMs, err := strconv.Atoi(publishTimeoutMsStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to read GENERATOR_PUBLISH_TIMEOUT_MS env variable: %s", err))
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

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
		go device.PublishData(ctx, wg)
	}

	keepRunning := true
	for keepRunning {
		select {
		case <-ctx.Done():
			fmt.Println("Terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			fmt.Println("Terminating: via signal")
			keepRunning = false
		}
	}
	cancel()
	wg.Wait()

	fmt.Println("### End of MQTT synthetic generation")
}

func ReadEnvVariable(env, fallback string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		return fallback
	}
	return v
}