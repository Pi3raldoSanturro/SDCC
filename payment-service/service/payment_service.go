package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"payment-service/messaging"
	"payment-service/models"
	"payment-service/proto"
	"payment-service/repository"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type PaymentService struct {
	Repo          *repository.PaymentRepository
	RabbitChannel *amqp091.Channel
	payment.UnimplementedPaymentServiceServer
}

func (s *PaymentService) Ping(ctx context.Context, req *payment.PingRequest) (*payment.PingResponse, error) {
	return &payment.PingResponse{Message: "Payment Service Pong!"}, nil
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	transactionID := uuid.New().String()

	paymentRecord := &models.Payment{
		UserID:          req.UserId,
		Amount:          req.Amount,
		Success:         true,
		TransactionId:   transactionID,
		EventInstanceId: "", // non usato qui, solo nei messaggi RabbitMQ
	}

	err := s.Repo.CreatePayment(ctx, paymentRecord)
	if err != nil {
		return nil, err
	}

	return &payment.PaymentResponse{
		Success:       true,
		TransactionId: transactionID,
		Message:       "Payment processed successfully",
	}, nil
}

func (s *PaymentService) StartConsumingTicketReservedEvents() {
	go func() {
		for {
			conn := messaging.ConnectRabbitMQWithRetry()
			ch := messaging.CreateChannel(conn)
			messaging.DeclareQueue(ch, "ticket-reserved-queue")

			closeChan := make(chan *amqp091.Error)
			ch.NotifyClose(closeChan)

			msgs, err := messaging.SubscribeToQueue(ch, "ticket-reserved-queue")
			if err != nil {
				log.Printf("Errore subscribe ticket-reserved-queue: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}

			log.Println("In ascolto su 'ticket-reserved-queue'...")
		consumeLoop:
			for {
				select {
				case d, ok := <-msgs:
					if !ok {
						log.Println("Canale chiuso. Restart consumer...")
						break consumeLoop
					}

					var event models.TicketReservedEvent
					if err := json.Unmarshal(d.Body, &event); err != nil {
						log.Printf("Errore decoding evento: %v", err)
						continue
					}

					//De-duplicazione basata su EventInstanceId
					duplicate, err := s.Repo.ExistsByEventInstanceID(context.Background(), event.EventInstanceId)
					if err != nil {
						log.Printf("Errore controllo duplicato: %v", err)
						continue
					}
					if duplicate {
						log.Printf("Evento duplicato ignorato: %s", event.EventInstanceId)
						continue
					}

					log.Printf("Evento da processare: %+v", event)

					transactionID := uuid.New().String()
					paymentSucceeded := true

					log.Printf("Simulazione pagamento: %v", paymentSucceeded)

					paymentRecord := &models.Payment{
						UserID:          event.UserId,
						Amount:          event.TotalAmount,
						Success:         paymentSucceeded,
						TransactionId:   transactionID,
						EventInstanceId: event.EventInstanceId,
					}

					if err := s.Repo.CreatePayment(context.Background(), paymentRecord); err != nil {
						log.Printf("Errore salvataggio pagamento: %v", err)
						continue
					}

					paymentEvent := map[string]interface{}{
						"eventId":         event.EventTicketId,
						"userId":          event.UserId,
						"success":         paymentSucceeded,
						"quantity":        event.Quantity,
						"eventInstanceId": event.EventInstanceId,
					}

					if err := messaging.PublishMessage("payment-events-queue", paymentEvent); err != nil {
						log.Printf("Errore pubblicazione PaymentEvent: %v", err)
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
