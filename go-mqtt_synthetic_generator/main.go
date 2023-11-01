package main

import (
	"fmt"
	"strconv"
	"sync"
)

func main() {
	// https://github.com/eclipse/paho.mqtt.golang
	//var broker = "tcp://test.mosquitto.org:1883"
	const broker = "andromeda.lasdpc.icmc.usp.br:31883"
	const topic = "topic/secret"
	const numDevices = 5000
	const deviceNumEvents = 100000

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
