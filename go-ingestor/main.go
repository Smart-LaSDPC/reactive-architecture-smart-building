package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go-ingestor/config"
	"go-ingestor/data"
	"go-ingestor/database"
	"go-ingestor/kafka"
)

func main() {
	// Config
	appConfig, err := config.GetAppConfig()
	if err != nil {
		log.Panicf("Failed to read app configuration: %s", err)
	}

	saramaConfig, err := kafka.GetSaramaConfig(appConfig)
	if err != nil {
		log.Panicf("Failed to generate Sarama library configuration: %s", err)
	}

	// Clientes
	kafkaClient, err := sarama.NewConsumerGroup([]string{appConfig.Kafka.BrokerAddress}, appConfig.Kafka.ConsumerGroupID, saramaConfig)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	repository, err := database.NewRepository(appConfig)
	if err != nil {
		log.Panicf("Failed create database client: %s", err)
	}

	// Metricas
	metricMessagesReceived := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ingestor_messages_received",
		Help: "Messages consumed from Kafka topic by Ingestor",
	})
	metricMessagesProcessed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ingestor_messages_processed",
		Help: "Messages received and proccessed by Ingestor",
	})
	metricInsertedSuccess := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ingestor_messages_inserted_success",
		Help: "Messages succesfully inserted into database by Ingestor",
	})
	metricInsertedFailed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ingestor_messages_inserted_fail",
		Help: "Messages failed to be inserted into database by Ingestor",
	})

	prometheus.MustRegister(
		metricMessagesReceived,
		metricMessagesProcessed,
		metricInsertedSuccess,
		metricInsertedFailed,
	)

	go func(){
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
		log.Printf("Started metrics server at :8080/metrics\n")
	}()

	// Fluxo de execucao
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	messages := consumeMessages(kafkaClient, appConfig, ctx, wg, metricMessagesReceived)

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
			go processAndInsert(ctx, wg, repository, appConfig, msg, metricMessagesProcessed, metricInsertedSuccess, metricInsertedFailed)
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

func processAndInsert(ctx context.Context, wg *sync.WaitGroup, repository *database.Repository, appConfig *config.AppConfig, msgBytes []byte, processedCounter, insertionSuccessCounter, insertionFailCounter prometheus.Counter) {
	defer wg.Done()

	msg, err := data.ParseMessageData(msgBytes)
	if err != nil {
		log.Printf("Failed to parse message: %s", err)
		return
	}
	log.Printf("Parsed message: %+v\n", msg)

	processedCounter.Inc()

	err = repository.InsertMsg(ctx, appConfig, msg)
	if err != nil {
		log.Printf("Failed to write message to database: %+v: %s", msg, err)
		insertionFailCounter.Inc()
		return
	}
	insertionSuccessCounter.Inc()
}

func consumeMessages(client sarama.ConsumerGroup, appConfig *config.AppConfig, ctx context.Context, wg *sync.WaitGroup, receivedCounter prometheus.Counter) chan []byte {
	consumer := &kafka.Consumer{
		Ready:           make(chan bool),
		Received:        make(chan []byte),
		ReceivedCounter: receivedCounter,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(consumer.Received)
		for {
			err := client.Consume(ctx, []string{appConfig.Kafka.Topic}, consumer)
			if err != nil {
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
