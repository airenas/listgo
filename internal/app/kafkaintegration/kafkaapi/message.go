package kafkaapi;

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	
)

// Msg wrapped msg to be returned by kafka
type Msg struct {
	ID string
	Offset kafka.TopicPartition
}