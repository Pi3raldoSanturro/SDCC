package messaging

import (
	"github.com/rabbitmq/amqp091-go"
)

func SubscribeToQueue(channel *amqp091.Channel, queueName string) (<-chan amqp091.Delivery, error) {
	msgs, err := channel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
