package kafka

import (
	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

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
	return &res, nil
}

//Write writes msg to Kafka
func (sp *Writer) Write(msg *kafkaapi.ResponseMsg) error {
	deliveryChan := make(chan ckafka.Event)
	defer close(deliveryChan)

	value := "Hello Go!"
	err := sp.producer.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &sp.topic, Partition: ckafka.PartitionAny},
		Value:          []byte(value),
		Headers:        []ckafka.Header{}}, deliveryChan)
	if err != nil {
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
