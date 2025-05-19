/*
package main

import (

	"context"
	"log"
	"net"

	"payment-service/messaging"
	pb "payment-service/proto"
	"payment-service/repository"
	"payment-service/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

)

	func connectMongo() *mongo.Database {
		clientOptions := options.Client().ApplyURI("mongodb://mongo-payment:27017")
		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			log.Fatalf("MongoDB connection error: %v", err)
		}

		if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
			log.Fatalf("MongoDB ping error: %v", err)
		}

		return client.Database("paymentdb")
	}

	func main() {
		lis, err := net.Listen("tcp", ":50053")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		db := connectMongo()
		repo := repository.NewPaymentRepository(db)

		// Connect RabbitMQ
		rabbitConn := messaging.ConnectRabbitMQ()
		rabbitChannel := messaging.CreateChannel(rabbitConn)
		messaging.DeclareQueue(rabbitChannel, "ticket-reserved-queue")
		messaging.DeclareQueue(rabbitChannel, "payment-events-queue")

		srv := grpc.NewServer()

		// 🔥 Creiamo l'istanza vera
		svc := &service.PaymentService{
			Repo:          repo,
			RabbitChannel: rabbitChannel,
		}

		pb.RegisterPaymentServiceServer(srv, svc)

		// 🔥 Start listener RabbitMQ
		svc.StartConsumingTicketReservedEvents()

		reflection.Register(srv)

		log.Println("Payment Service is running on port 50053")
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}
*/

/*
package main

import (
	"context"
	"log"
	"net"

	"payment-service/messaging"
	pb "payment-service/proto"
	"payment-service/repository"
	"payment-service/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func connectMongo() *mongo.Database {
	clientOptions := options.Client().ApplyURI("mongodb://mongo-payment:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatalf("❌ MongoDB ping error: %v", err)
	}

	log.Println("✅ Connessione a MongoDB (payment-service) riuscita.")
	return client.Database("paymentdb")
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("❌ failed to listen: %v", err)
	}

	db := connectMongo()
	repo := repository.NewPaymentRepository(db)
	/*
		// 🔌 Connessione con retry a RabbitMQ
		rabbitConn := messaging.ConnectRabbitMQWithRetry()
		rabbitChannel := messaging.CreateChannel(rabbitConn)

		// 📦 Dichiara le code usate
		messaging.DeclareQueue(rabbitChannel, "ticket-reserved-queue")
		messaging.DeclareQueue(rabbitChannel, "payment-events-queue")
*/
/*
	// ✅ Inizializza CircuitBreaker
	messaging.GlobalBreaker = messaging.NewCircuitBreaker("amqp://guest:guest@rabbitmq:5672/", []string{
		"ticket-reserved-queue",
		"payment-events-queue",
	})

	if messaging.GlobalBreaker == nil || messaging.GlobalBreaker.Channel == nil {
		log.Fatal("❌ Circuit breaker non inizializzato correttamente")
	}

	// ⚙️ Inizializza servizio gRPC
	srv := grpc.NewServer()
	svc := &service.PaymentService{
		Repo: repo,
	}
	pb.RegisterPaymentServiceServer(srv, svc)

	// 🔁 Inizia a consumare eventi RabbitMQ con recovery
	svc.StartConsumingTicketReservedEvents()

	reflection.Register(srv)

	log.Println("🚀 Payment Service is running on port 50053")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("❌ failed to serve: %v", err)
	}
}
*/
package main

import (
	"context"
	"log"
	"net"
	"time"

	"payment-service/messaging"
	pb "payment-service/proto"
	"payment-service/repository"
	"payment-service/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func connectMongo() *mongo.Database {
	clientOptions := options.Client().ApplyURI("mongodb://mongo-payment:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatalf("❌ MongoDB ping error: %v", err)
	}

	log.Println("✅ Connessione a MongoDB (payment-service) riuscita.")
	return client.Database("paymentdb")
}

func waitForRabbitAndInitBreaker() {
	const maxRetries = 30
	for i := 1; i <= maxRetries; i++ {
		messaging.GlobalBreaker = messaging.NewCircuitBreaker("amqp://guest:guest@rabbitmq:5672/", []string{
			"ticket-reserved-queue",
			"payment-events-queue",
		})

		if messaging.GlobalBreaker != nil && messaging.GlobalBreaker.Channel != nil {
			log.Println("✅ Circuit breaker inizializzato correttamente (payment-service).")
			return
		}

		log.Printf("⏳ RabbitMQ non ancora pronto (tentativo %d/%d). Riprovo in 2s...", i, maxRetries)
		time.Sleep(2 * time.Second)
	}

	log.Fatal("❌ Impossibile inizializzare il Circuit Breaker: RabbitMQ non disponibile dopo vari tentativi.")
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("❌ failed to listen: %v", err)
	}

	db := connectMongo()
	repo := repository.NewPaymentRepository(db)

	// ⏳ Aspetta RabbitMQ e inizializza Circuit Breaker
	waitForRabbitAndInitBreaker()

	// ⚙️ Inizializza servizio gRPC
	srv := grpc.NewServer()
	svc := &service.PaymentService{
		Repo: repo,
	}
	pb.RegisterPaymentServiceServer(srv, svc)

	// 🔁 Inizia a consumare eventi RabbitMQ con recovery
	svc.StartConsumingTicketReservedEvents()

	reflection.Register(srv)

	log.Println("🚀 Payment Service is running on port 50053")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("❌ failed to serve: %v", err)
	}
}
