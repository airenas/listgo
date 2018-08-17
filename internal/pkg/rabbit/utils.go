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
