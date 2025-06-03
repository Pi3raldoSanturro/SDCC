package main

import (
	"context"
	"log"
	"net"
	"time"

	"ticket-service/messaging"
	"ticket-service/proto"
	"ticket-service/repository"
	"ticket-service/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func connectMongo() *mongo.Database {
	clientOptions := options.Client().ApplyURI("mongodb://mongo-ticket:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}

	log.Println("Connessione a MongoDB (ticket-service) riuscita.")
	return client.Database("ticketdb")
}

func waitForRabbitAndInitBreaker() {
	const maxRetries = 30
	for i := 1; i <= maxRetries; i++ {
		messaging.GlobalBreaker = messaging.NewCircuitBreaker("amqp://guest:guest@rabbitmq:5672/", []string{
			"ticket-reserved-queue",
			"payment-events-queue",
		})

		if messaging.GlobalBreaker != nil && messaging.GlobalBreaker.Channel != nil {
			log.Println("Circuit breaker inizializzato correttamente.")
			return
		}

		log.Printf("RabbitMQ non ancora pronto (tentativo %d/%d). Riprovo in 2s...", i, maxRetries)
		time.Sleep(2 * time.Second)
	}

	log.Fatal("Impossibile inizializzare il Circuit Breaker: RabbitMQ non disponibile dopo vari tentativi.")
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db := connectMongo()
	repo := repository.NewEventRepository(db)

	waitForRabbitAndInitBreaker()

	messaging.StartConsumingPaymentEvents(repo)

	srv := grpc.NewServer()
	ticket.RegisterTicketServiceServer(srv, &service.TicketService{
		Repo: repo,
	})
	reflection.Register(srv)

	log.Println("Ticket Service is running on port 50052")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
