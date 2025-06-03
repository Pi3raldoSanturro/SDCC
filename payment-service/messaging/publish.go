package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var GlobalBreaker *CircuitBreaker

func PublishMessage(queueName string, message interface{}) error {
	if GlobalBreaker == nil || GlobalBreaker.Channel == nil {
		log.Println("Circuit breaker non inizializzato o channel nullo")
		return fmt.Errorf("circuit breaker non inizializzato")
	}

	if GlobalBreaker.IsOpen() {
		log.Printf("Circuito aperto: impossibile pubblicare su '%s'", queueName)
		return fmt.Errorf("circuit breaker aperto")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	for i := 1; i <= 5; i++ {
		err = GlobalBreaker.Channel.Publish(
			"",
			queueName,
			false,
			false,
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
		if err == nil {
			log.Printf("Messaggio pubblicato su '%s'", queueName)
			return nil
		}

		log.Printf("Pubblicazione fallita su '%s' (tentativo %d/5): %v", queueName, i, err)
		time.Sleep(2 * time.Second)
	}

	GlobalBreaker.MarkFailure()
	if GlobalBreaker.TryRecover() {
		log.Println("Riconnessione riuscita")
	} else {
		log.Println("Riconnessione fallita")
	}

	return fmt.Errorf("publish failed after retries: %w", err)
}
