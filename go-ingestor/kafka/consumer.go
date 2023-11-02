package kafka

import (
	"log"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"

)

type Consumer struct {
	Ready chan bool
	Received chan []byte
	ReceivedCounter prometheus.Counter
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.Ready)
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("Message channel was closed")
				return nil
			}
			session.MarkMessage(message, "")
			consumer.ReceivedCounter.Inc()
			consumer.Received <- message.Value
		case <-session.Context().Done():
			return nil
		}
	}
}
