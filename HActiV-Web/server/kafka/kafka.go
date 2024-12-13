package kafka

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/beego/beego/v2/core/logs"
)

var (
	producer  sarama.SyncProducer
	consumer  sarama.Consumer
	consumers map[string]sarama.PartitionConsumer
	mu        sync.Mutex
	MessageChannel chan []byte
)

func init() {
	consumers = make(map[string]sarama.PartitionConsumer)
	MessageChannel = make(chan []byte, 1000)
}

func InitKafka(brokers []string) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	var err error
	producer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	consumer, err = sarama.NewConsumer(brokers, config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return nil
}

func ProduceMessage(topic string, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message value: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(jsonValue),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	logs.Info("Message successfully sent to topic: %s, partition: %d, offset: %d", topic, partition, offset)
	return nil
}

func ConsumeMessages(topic string, handler func([]byte) error) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := consumers[topic]; exists {
		return fmt.Errorf("consumer for topic %s already exists", topic)
	}

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions for topic %s: %w", topic, err)
	}

	for _, partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to create partition consumer: %w", err)
		}

		consumers[fmt.Sprintf("%s-%d", topic, partition)] = partitionConsumer

		go func(pc sarama.PartitionConsumer) {
			for {
				select {
				case msg := <-pc.Messages():
					logs.Debug("Received message from topic %s: %s", topic, string(msg.Value))
					
					if err := handler(msg.Value); err != nil {
						logs.Error("Error handling message: %v", err)
					}

					// Send message to channel for WebSocket broadcasting
					select {
					case MessageChannel <- msg.Value:
						logs.Debug("Message sent to WebSocket channel")
					default:
						logs.Warn("WebSocket channel buffer full, message dropped")
					}

				case err := <-pc.Errors():
					logs.Error("Error consuming message from topic %s: %v", topic, err)
				}
			}
		}(partitionConsumer)
	}

	logs.Info("Started consuming messages from topic: %s", topic)
	return nil
}

func GetMessageChannel() chan []byte {
	return MessageChannel
}

func CloseKafka() {
	if producer != nil {
		if err := producer.Close(); err != nil {
			logs.Error("Error closing Kafka producer: %v", err)
		} else {
			logs.Info("Kafka producer closed successfully")
		}
	}

	for topic, partitionConsumer := range consumers {
		if err := partitionConsumer.Close(); err != nil {
			logs.Error("Error closing Kafka consumer for topic %s: %v", topic, err)
		} else {
			logs.Info("Kafka consumer for topic %s closed successfully", topic)
		}
	}

	if consumer != nil {
		if err := consumer.Close(); err != nil {
			logs.Error("Error closing Kafka consumer: %v", err)
		} else {
			logs.Info("Kafka consumer closed successfully")
		}
	}

	close(MessageChannel)
}

