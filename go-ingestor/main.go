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
	"time"

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
		metricInsertedSuccess,
		metricInsertedFailed,
	)

	go func () {
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

	batchSize := appConfig.DB.BatchSize
	flushTimeout := appConfig.DB.BatchFlushTimeout

	msgBatch := make([][]byte, batchSize)

	lastWrittenTime := time.Now()
	currBatchSize := 0

	flushMessages := make(chan bool)
	go func(){
		for {
			if time.Since(lastWrittenTime) > flushTimeout {
				lastWrittenTime = time.Now()
				flushMessages <- true
			}
		}
	}()

	keepRunning := true
	for keepRunning {
		select {
		case <-flushMessages:
			if currBatchSize > 0 {
				wg.Add(1)
				go processBatch(ctx, wg, repository, appConfig, msgBatch[:currBatchSize], metricInsertedSuccess, metricInsertedFailed)
				currBatchSize = 0
			}
		case msg, ok := <-messages:
			if !ok {
				log.Println("Terminating: done receiving messages")
				keepRunning = false
				break
			}
			msgBatch[currBatchSize] = msg
			currBatchSize += 1

			if currBatchSize == batchSize {
				wg.Add(1)
				go processBatch(ctx, wg, repository, appConfig, msgBatch, metricInsertedSuccess, metricInsertedFailed)
				lastWrittenTime = time.Now()
				currBatchSize = 0
			}
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

func processBatch(ctx context.Context, wg *sync.WaitGroup, repository *database.Repository, appConfig *config.AppConfig, msgBatch [][]byte, insertionSuccessCounter, insertionFailCounter prometheus.Counter) {
	defer wg.Done()
	var err error

	msgs := []data.MessageData{}
	for _, msgRaw := range msgBatch {
		msg, err := data.ParseMessageData(msgRaw)
		if err != nil {
			log.Printf("Failed to parse message: %s", err)
			return
		}
		msgs = append(msgs, *msg)
	} 

	err = repository.InsertMsgBatch(ctx, appConfig, msgs)
	if err != nil {
		log.Printf("Failed to write message batch to database: %s", err)
		insertionFailCounter.Add( float64(len(msgBatch)))
		return
	}
	log.Printf("Successfully inserted message batch into database\n")
	insertionSuccessCounter.Add(float64(len(msgBatch)))
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
