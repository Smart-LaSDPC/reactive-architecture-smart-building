package main

import (
	"fmt"
	"strconv"
	"sync"
)

func main() {
	// https://github.com/eclipse/paho.mqtt.golang
	//var broker = "tcp://test.mosquitto.org:1883"
	// "andromeda.lasdpc.icmc.usp.br:31883"
	const broker = "localhost:1883"
	const topic = "farm_id/device_id/testclient1"
	const numDevices = 40
	const deviceNumEvents = 200
	// simulates device subscription
	var wg sync.WaitGroup

	for i := 0; i < numDevices; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub_generate(strconv.Itoa(i), broker, topic, deviceNumEvents)
			fmt.Println("Device goroutine scheduled")
		}()

	}
	wg.Wait()

	fmt.Println("### End of MQTT synthetic generation")
}
