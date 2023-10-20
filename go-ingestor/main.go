package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"go-ingestor/config"
	"go-ingestor/data"
	"go-ingestor/kafka"

	"github.com/IBM/sarama"

	_ "github.com/lib/pq"
)

const (
	host     = "!" // Endereço do banco de dados PostgreSQL
	port     = !            // Porta padrão do PostgreSQL
	user     = "!"         // Seu nome de usuário do PostgreSQL
	password = "!"         // Sua senha do PostgreSQL
	dbname   = "!"      // Nome do banco de dados
)

// como lidar com multiplas replicas e particoes
// salvar offset da ultima mensagem lida ao desconectar

func main() {
	appConfig, err := config.GetAppConfig()
	if err != nil {
		log.Panicf("Failed to read app configuration: %s", err)
	}

	saramaConfig, err := kafka.GetSaramaConfig(appConfig)
	if err != nil {
		log.Panicf("Failed to generate Sarama library configuration: %s", err)
	}

	consumer := kafka.Consumer{
		Ready:    make(chan bool),
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
			if err := client.Consume(ctx, []string{appConfig.Kafka.Topic}, &consumer); err != nil {
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
	log.Printf("Consumer running for messages on %s topic", appConfig.Kafka.Topic)

	wg.Add(1)
	go func(ctx context.Context, cancel context.CancelFunc) {
		defer wg.Done()

		select {
		case msgBytes, ok := <-consumer.Received:
			if !ok {
				return
			}
			msg, err := data.ParseMessageData(msgBytes)
			if err != nil {
				log.Printf("Failed reading message: %s", err)
			}
			log.Printf("%+v", msg)
			insertData(*msg)
		case <-ctx.Done():
			return
		}
	}(ctx, cancel)

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

func insertData(msg data.MessageData) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO tbl_temperature_moisture (time, agent_id, state, temperature, moisture) VALUES ($1, $2, $3, $4, $5)", msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Dados inseridos com sucesso no PostgreSQL.")
	}
}
