/*
package messaging

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

func PublishMessage(channel *amqp091.Channel, queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = channel.Publish(
		"",        // no exchange (default direct)
		queueName, // routing key = nome coda
		false,     // mandatory
		false,     // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}

	log.Printf("Published message to queue: %s", queueName)
	return nil
}
*/
/*
package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func PublishMessage(channel *amqp091.Channel, queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

	for i := 1; i <= 5; i++ {
		err = channel.Publish(
			"",        // no exchange
			queueName, // routing key
			false,
			false,
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
		if err == nil {
			log.Printf("✅ Published message to queue '%s'", queueName)
			return nil
		}

		log.Printf("❌ Failed to publish message to queue '%s' (attempt %d/5): %v", queueName, i, err)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("publish failed after retries: %w", err)
}
*/
package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

var GlobalBreaker *CircuitBreaker

func PublishMessage(queueName string, message interface{}) error {
	if GlobalBreaker == nil {
		log.Println("❌ GlobalBreaker è nil!")
		return fmt.Errorf("GlobalBreaker non inizializzato")
	}
	if GlobalBreaker.Channel == nil {
		log.Println("❌ Channel è nil al momento della pubblicazione")
	}

	if GlobalBreaker.IsOpen() {
		log.Printf("🛑 Circuito aperto. Impossibile pubblicare su '%s'", queueName)
		return fmt.Errorf("circuit breaker open")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

	for i := 1; i <= 5; i++ {
		if GlobalBreaker.Channel == nil {
			log.Printf("⚠️ Channel non disponibile (tentativo %d/5)", i)
			time.Sleep(2 * time.Second)
			continue
		}

		err = GlobalBreaker.Channel.Publish(
			"",        // no exchange
			queueName, // routing key
			false,
			false,
			NewPublishing(body),
		)
		if err == nil {
			log.Printf("✅ Messaggio pubblicato su '%s'", queueName)
			return nil
		}

		log.Printf("❌ Pubblicazione fallita su '%s' (tentativo %d/5): %v", queueName, i, err)
		if i == 5 {
			GlobalBreaker.MarkFailure()
		}
		time.Sleep(2 * time.Second)
	}

	if !GlobalBreaker.IsOpen() {
		return fmt.Errorf("publish failed: %w", err)
	}

	if GlobalBreaker.TryRecover() {
		log.Println("🔌 Riconnessione riuscita. Puoi riprovare la pubblicazione.")
	} else {
		log.Println("❌ Riconnessione fallita.")
	}

	return fmt.Errorf("publish failed after retries: %w", err)
}

func NewPublishing(body []byte) amqp091.Publishing {
	return amqp091.Publishing{
		ContentType: "application/json",
		Body:        body,
	}
}
