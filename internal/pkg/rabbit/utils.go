package rabbit

import "github.com/streadway/amqp"

//DeclareQueue decrares durable queue
func DeclareQueue(ch *amqp.Channel, qName string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		qName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

//NewChannel creates channel to listen from rabbit with auto ack = false
func NewChannel(ch *amqp.Channel, qName string) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		qName, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

//DeclareExchange creates exchange to publish events
func DeclareExchange(ch *amqp.Channel, topic string) error {
	return ch.ExchangeDeclare(
		topic,    // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
}
