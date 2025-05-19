package messaging

import (
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type HandlerFunc func(body []byte)

func StartConsumingWithRecovery(queueName string, handler HandlerFunc) {
	go func() {
		for {
			conn := ConnectRabbitMQWithRetry()
			channel := CreateChannel(conn)
			DeclareQueue(channel, queueName)

			notifyClose := make(chan *amqp091.Error)
			channel.NotifyClose(notifyClose)

			msgs, err := SubscribeToQueue(channel, queueName)
			if err != nil {
				log.Printf("❌ Subscribe error on queue '%s': %v", queueName, err)
				time.Sleep(3 * time.Second)
				continue
			}

			log.Printf("📥 Listening on queue '%s'...", queueName)
		consumeLoop:
			for {
				select {
				case d, ok := <-msgs:
					if !ok {
						log.Println("⚠️ Channel closed. Restarting consumer...")
						break consumeLoop
					}
					handler(d.Body)
				case <-notifyClose:
					log.Println("⚠️ Channel notifyClose triggered. Restarting consumer...")
					break consumeLoop
				}
			}

			time.Sleep(3 * time.Second)
		}
	}()
}

/*
func ConnectRabbitMQWithRetry() *amqp091.Connection {
	var conn *amqp091.Connection
	var err error
	for i := 1; i <= 5; i++ {
		conn, err = amqp091.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			log.Println("✅ RabbitMQ connection established")
			return conn
		}
		log.Printf("❌ Failed to connect to RabbitMQ (attempt %d/5): %v", i, err)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("❌ Could not connect to RabbitMQ after retries: %v", err)
	return nil
}
*/

func ConnectRabbitMQWithRetry() *amqp091.Connection {
	var conn *amqp091.Connection
	var err error
	const maxRetries = 30
	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp091.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			log.Println("✅ Connessione a RabbitMQ riuscita.")
			return conn
		}
		log.Printf("⏳ Tentativo %d/%d: RabbitMQ non disponibile. Riprovo in 2s...", i, maxRetries)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("❌ RabbitMQ non disponibile dopo %d tentativi: %v", maxRetries, err)
	return nil
}
