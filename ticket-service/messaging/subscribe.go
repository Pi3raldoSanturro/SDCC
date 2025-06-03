package messaging

import (
	"context"
	"encoding/json"
	"log"
	"ticket-service/repository"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type PaymentEvent struct {
	EventId         string `json:"eventId"`
	Success         bool   `json:"success"`
	UserId          string `json:"userId"`
	Quantity        int32  `json:"quantity"`
	EventInstanceId string `json:"eventInstanceId"`
}

func StartConsumingPaymentEvents(repo *repository.EventRepository) {
	go func() {
		for {
			conn := ConnectRabbitMQWithRetry()
			ch := CreateChannel(conn)
			DeclareQueue(ch, "payment-events-queue")

			closeChan := make(chan *amqp091.Error)
			ch.NotifyClose(closeChan)

			msgs, err := SubscribeToQueue(ch, "payment-events-queue")
			if err != nil {
				log.Printf("Errore subscribe: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}

			log.Println("In ascolto sulla coda 'payment-events-queue'...")
		consumeLoop:
			for {
				select {
				case d, ok := <-msgs:
					if !ok {
						log.Println("Canale chiuso. Riavvio consumer...")
						break consumeLoop
					}

					var event PaymentEvent
					if err := json.Unmarshal(d.Body, &event); err != nil {
						log.Printf("Errore parsing evento pagamento: %v", err)
						continue
					}

					// ✅ Controlla duplicati
					alreadyProcessed, err := repo.HasProcessedEvent(context.Background(), event.EventInstanceId)
					if err != nil {
						log.Printf("Errore controllo duplicato pagamento: %v", err)
						continue
					}
					if alreadyProcessed {
						log.Printf("Evento duplicato ignorato: %s", event.EventInstanceId)
						continue
					}

					if event.Success {
						log.Printf("Pagamento riuscito per evento %s", event.EventId)
						repo.MarkTicketAsPaid(context.Background(), event.EventId)
					} else {
						log.Printf("Pagamento fallito per %s. Ripristino %d biglietti", event.EventId, event.Quantity)
						repo.RestoreTickets(context.Background(), event.EventId, event.Quantity)
					}

					// ✅ Marca evento come processato
					if err := repo.MarkEventAsProcessed(context.Background(), event.EventInstanceId); err != nil {
						log.Printf("Impossibile salvare EventInstanceId %s come processato: %v", event.EventInstanceId, err)
					}
				case <-closeChan:
					log.Println("Connessione chiusa da RabbitMQ. Restart consumer...")
					break consumeLoop
				}
			}

			time.Sleep(3 * time.Second)
		}
	}()
}

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
