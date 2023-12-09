package kafka

import (
	"encoding/json"
	"os"

	"github.com/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

// Reader reads data from kafka topic
type Reader struct {
	consumer    *ckafka.Consumer
	stopChannel <-chan os.Signal
}

// NewReader creates Kafka reader
func NewReader(stopChannel <-chan os.Signal) (*Reader, error) {
	brokers := cmdapp.Config.GetString("kafka.brokers")
	if brokers == "" {
		return nil, errors.New("No kafka.brokers provided")
	}
	group := cmdapp.Config.GetString("kafka.group")
	if group == "" {
		group = "Transcriber.Service.Group"
	}
	topic := cmdapp.Config.GetString("kafka.inputTopic")
	if topic == "" {
		return nil, errors.New("No kafka.inputTopic provided")
	}

	res := Reader{}
	res.stopChannel = stopChannel

	//sessionTimeout := 30 * time.Minute

	cmdapp.Log.Infof("Connecting to Kafka on %s, group: %s", brokers, group)
	var err error
	res.consumer, err = ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers":     brokers,
		"broker.address.family": "v4",
		"group.id":              group,
		"session.timeout.ms":    6000, //0int(sessionTimeout.Seconds()) * 1000,
		//"max.poll.interval.ms":  int(sessionTimeout.Seconds()) * 1000,
		"auto.offset.reset":  "latest",
		"enable.auto.commit": "false",
	})

	if err != nil {
		return nil, errors.Wrap(err, "Can't connect to kafka brokers: "+brokers)
	}
	cmdapp.Log.Infof("Subscribing to Kafka topic %s", topic)
	err = res.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Can't subscribe to kafka topic: "+topic)
	}
	cmdapp.Log.Infof("Subscribed to Kafka topic %s", topic)
	return &res, nil
}

// Close closes reader
func (sp *Reader) Close() {
	if sp.consumer != nil {
		cmdapp.Log.Info("Closing kafka consumer")
		sp.consumer.Close()
		cmdapp.Log.Info("Closed kafka consumer")
	}
}

// Commit commit messages offset
func (sp *Reader) Commit(msg *kafkaapi.Msg) error {
	cmdapp.Log.Infof("Commit message %s", msg.RealMsg.TopicPartition.String())
	_, err := sp.consumer.CommitMessage(msg.RealMsg)
	return err
}

// Get reads a next message from kafka topic
func (sp *Reader) Get() (*kafkaapi.Msg, error) {
	for {
		select {
		case <-sp.stopChannel:
			break
		default:
			ev := sp.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *ckafka.Message:
				cmdapp.Log.Debugf("Kafka message on %s:\t%s\n", e.TopicPartition, string(e.Value))
				id, err := parseMsg(e.Value)
				if err != nil {
					return nil, errors.Wrapf(err, "Cannot parse json %s", string(e.Value))
				}
				msg := kafkaapi.Msg{}
				msg.ID = id
				msg.RealMsg = e
				return &msg, nil
			case ckafka.Error:
				cmdapp.Log.Warnf("Kafka warning: %v: %v\n", e.Code(), e)
				if e.Code() == ckafka.ErrAllBrokersDown {
					return nil, errors.New("All kafka brokers down")
				}
			default:
				cmdapp.Log.Debugf("Ignored kafka event type %v\n", e)
			}
		}
	}
}

type kafkaMsg struct {
	ID string `json:"id"`
}

func parseMsg(d []byte) (string, error) {
	var res kafkaMsg
	err := json.Unmarshal(d, &res)
	if err != nil {
		return "", err
	}
	return res.ID, nil
}
