package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go-ingestor/config"
	"go-ingestor/data"
	"go-ingestor/kafka"
	"go-ingestor/database"

	"github.com/IBM/sarama"

	_ "github.com/lib/pq"
)

func main() {
	appConfig, err := config.GetAppConfig()
	if err != nil {
		log.Panicf("Failed to read app configuration: %s", err)
	}

	saramaConfig, err := kafka.GetSaramaConfig(appConfig)
	if err != nil {
		log.Panicf("Failed to generate Sarama library configuration: %s", err)
	}
	
	kafkaClient, err := sarama.NewConsumerGroup([]string{appConfig.Kafka.BrokerAddress}, appConfig.Kafka.ConsumerGroupID, saramaConfig)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	repository, err := database.NewRepository(appConfig)
	if err != nil {
		log.Panicf("Failed create database client: %s", err)
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	messages := consumeMessages(kafkaClient, appConfig, ctx, wg)
	
	keepRunning := true
	for keepRunning {
		select {
		case msg, ok := <-messages:
			if !ok {
				log.Println("Terminating: done receiving messages")
				keepRunning = false
				break
			}
			wg.Add(1)
			go processAndInsert(repository, appConfig, wg, msg)
		case <-ctx.Done():
			log.Println("Terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("Terminating: via signal")
			keepRunning = false
		}
	}

	cancel()
	wg.Wait()
	repository.Close()

	if err = kafkaClient.Close(); err != nil {
		log.Panicf("Error closing Sarama client: %v", err)
	}
}

func processAndInsert(repository *database.Repository, appConfig *config.AppConfig, wg *sync.WaitGroup, msgBytes []byte) {
	defer wg.Done()

	msg, err := data.ParseMessageData(msgBytes)
	if err != nil {
		log.Printf("Failed to parse message: %s", err)
		return
	}

	err = repository.InsertMsg(appConfig, msg)
	if err != nil {
		log.Printf("Failed to write message to database: %+v: %s", msg, err)
	}
}

func consumeMessages(client sarama.ConsumerGroup, appConfig *config.AppConfig, ctx context.Context, wg *sync.WaitGroup) (chan []byte) {
	consumer := &kafka.Consumer{
		Ready:    make(chan bool),
		Received: make(chan []byte),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(consumer.Received)
		for {
			if err := client.Consume(ctx, []string{appConfig.Kafka.Topic}, consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				log.Printf("Done consuming messages")
				return
			}
			consumer.Ready = make(chan bool)
			log.Printf("Consumer running for messages on %s topic", appConfig.Kafka.Topic)
		}
	}()
	<-consumer.Ready
	log.Printf("Started consuming messages from %s at topic %s", appConfig.Kafka.BrokerAddress, appConfig.Kafka.Topic)
	return consumer.Received
}
