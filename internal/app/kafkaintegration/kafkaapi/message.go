package kafkaapi;

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	
)

// Msg wrapped msg to be returned by kafka
type Msg struct {
	ID string
	RealMsg *kafka.Message
}

// ResponseMsg wrapped msg to be writen to kafkas AudioTextReadyEvent
type ResponseMsg struct {
	ID string
	Error TranscriptionError
}

// TranscriptionError keeps error structure to put into kafkas event
type TranscriptionError struct {
	Status string
	Msg string
}

