package kafka

import (
	"fmt"
	"log"
	"os"

	"github.com/IBM/sarama"

	"go-ingestor/config"
)

func GetSaramaConfig(appConfig *config.AppConfig) (*sarama.Config, error) {
	if appConfig.Kafka.Verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(appConfig.Kafka.Version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version

	switch appConfig.Kafka.AssignorStrategy {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		return nil, fmt.Errorf("Unrecognized consumer group partition assignor: %s", appConfig.Kafka.AssignorStrategy)
	}

	if appConfig.Kafka.OffsetOldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	return config, nil
}