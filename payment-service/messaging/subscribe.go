package messaging

import (
	"github.com/rabbitmq/amqp091-go"
)

func SubscribeToQueue(channel *amqp091.Channel, queueName string) (<-chan amqp091.Delivery, error) {
	msgs, err := channel.Consume(
		queueName,
		"",
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
