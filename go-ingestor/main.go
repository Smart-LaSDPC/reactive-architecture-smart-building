package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"log"

	"github.com/IBM/sarama"
	"go-ingestor/config"
	"go-ingestor/kafka"
)

// como lidar com multiplas replicas e particoes
// salvar offset da ultima mensagem lida ao desconectar

func main() {
	appConfig, err := config.GetAppConfig()
	if err != nil {
		log.Panicf("Failed to read app configuration: %s", err)
	} 

	saramaConfig, err := kafka.GenerateSaramaConfig(appConfig)
	if err != nil {
		log.Panicf("Failed to generate Sarama library configuration: %s", err)
	} 

	consumer := kafka.Consumer{
		Ready: make(chan bool),
		Received: make(chan []byte),
	}

	ctx, cancel := context.WithCancel(context.Background())

	client, err := sarama.NewConsumerGroup(strings.Split(appConfig.Kafka.Brokers, ","), appConfig.Kafka.ConsumerGroupID, saramaConfig)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(appConfig.Kafka.Topics, ","), &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.Ready = make(chan bool)
		}
	}()
	
	<-consumer.Ready
	log.Println("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	
	pause := false
	keepRunning := true
	
	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(client, &pause)
		}
	}

	cancel()
	wg.Wait()

	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}


func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		client.PauseAll()
		log.Println("Pausing consumption")
	}

	*isPaused = !*isPaused
}