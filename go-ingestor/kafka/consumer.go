package kafka

import (
	"log"
	"github.com/IBM/sarama"
)

type Consumer struct {
	Ready chan bool
	Received chan []byte
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
				log.Printf("message channel was closed")
				return nil
			}
			session.MarkMessage(message, "")
			consumer.Received <- message.Value
		case <-session.Context().Done():
			return nil
		}
	}
}