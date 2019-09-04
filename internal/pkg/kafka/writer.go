package kafka

import (
	"encoding/json"
	"fmt"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type kafkaRespMsg struct {
	ID    string             `json:"id"`
	Error *kafkaRespMsgError `json:"error,omitempty"`
}

type kafkaRespMsgError struct {
	Code         string `json:"code"`
	DebugMessage string `json:"debug_message"`
}

//Writer writes messages to Kafka topic
type Writer struct {
	producer *ckafka.Producer
	topic    string
}

//NewWriter creates Kafka writer
func NewWriter() (*Writer, error) {
	brokers := cmdapp.Config.GetString("kafka.brokers")
	if brokers == "" {
		return nil, errors.New("No kafka.brokers provided")
	}
	res := Writer{}
	res.topic = cmdapp.Config.GetString("kafka.resultTopic")
	if res.topic == "" {
		return nil, errors.New("No kafka.resultTopic provided")
	}

	cmdapp.Log.Infof("Connecting to Kafka on %s\n", brokers)
	var err error
	res.producer, err = ckafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return nil, errors.Wrap(err, "Can't connect to kafka brokers: "+brokers)
	}
	cmdapp.Log.Infof("Connected to Kafka on %s\n", brokers)
	return &res, nil
}

//Write writes msg to Kafka
func (sp *Writer) Write(msg *kafkaapi.ResponseMsg) error {
	deliveryChan := make(chan ckafka.Event)
	defer close(deliveryChan)

	kafkaMsg := newMessage(msg)
	value, err := json.Marshal(kafkaMsg)
	if err != nil {
		return errors.Wrap(err, "Can't marshal message before sending")
	}
	err = sp.producer.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &sp.topic, Partition: ckafka.PartitionAny},
		Value:          value,
		Headers:        []ckafka.Header{}}, deliveryChan)
	if err != nil {
		cmdapp.Log.Tracef("Msg: %s", string(value))
		return errors.Wrap(err, "Can't send message to kafka topic")
	}
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return errors.Wrap(m.TopicPartition.Error, "Can't deliver msg")
	}
	cmdapp.Log.Infof("Delivered message to topic %s [%d] at offset %v\n",
		*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	return nil
}

func newMessage(msg *kafkaapi.ResponseMsg) *kafkaRespMsg {
	res := kafkaRespMsg{}
	res.ID = msg.ID
	if msg.Error.Code != "" {
		res.Error = &kafkaRespMsgError{}
		res.Error.Code = msg.Error.Code
		res.Error.DebugMessage = checkTrimLen(msg.Error.DebugMessage, 20000)
	}
	return &res
}

func checkTrimLen(input string, l int) string {
	r := []rune(input)
	if len(r) > l {
		sr := fmt.Sprintf("Showing first %d symbols of the message\n", l)
		return sr + string(r[0:l]) + "..."
	}
	return input
}
