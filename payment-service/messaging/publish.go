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
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var GlobalBreaker *CircuitBreaker

func PublishMessage(queueName string, message interface{}) error {
	if GlobalBreaker == nil || GlobalBreaker.Channel == nil {
		log.Println("❌ Circuit breaker non inizializzato o channel nullo")
		return fmt.Errorf("circuit breaker non inizializzato")
	}

	if GlobalBreaker.IsOpen() {
		log.Printf("🛑 Circuito aperto: impossibile pubblicare su '%s'", queueName)
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
			log.Printf("✅ Messaggio pubblicato su '%s'", queueName)
			return nil
		}

		log.Printf("❌ Pubblicazione fallita su '%s' (tentativo %d/5): %v", queueName, i, err)
		time.Sleep(2 * time.Second)
	}

	GlobalBreaker.MarkFailure()
	if GlobalBreaker.TryRecover() {
		log.Println("🔌 Riconnessione riuscita")
	} else {
		log.Println("❌ Riconnessione fallita")
	}

	return fmt.Errorf("publish failed after retries: %w", err)
}
