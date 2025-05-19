/*
package main

import (

	"context"
	"log"
	"net"

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

		return client.Database("ticketdb")
	}

	func main() {
		lis, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		db := connectMongo()
		repo := repository.NewEventRepository(db)

		// Connect RabbitMQ
		rabbitConn := messaging.ConnectRabbitMQ()
		rabbitChannel := messaging.CreateChannel(rabbitConn)
		messaging.DeclareQueue(rabbitChannel, "ticket-reserved-queue")
		messaging.DeclareQueue(rabbitChannel, "payment-events-queue")

		srv := grpc.NewServer()

		ticket.RegisterTicketServiceServer(srv, &service.TicketService{
			Repo:          repo,
			RabbitChannel: rabbitChannel,
		})

		reflection.Register(srv)

		log.Println("Ticket Service is running on port 50052")
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}
*/

/*
package main

import (
	"context"
	"github.com/rabbitmq/amqp091-go"
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

	log.Println("✅ Connessione a MongoDB (ticket-service) riuscita.")
	return client.Database("ticketdb")
}

func connectRabbitMQWithRetry() *amqp091.Connection {
	var conn *amqp091.Connection
	var err error
	for i := 0; i < 5; i++ {
		conn, err = amqp091.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			return conn
		}
		log.Printf("Tentativo %d: RabbitMQ non disponibile, retry in 3s...", i+1)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("Impossibile connettersi a RabbitMQ dopo vari tentativi: %v", err)
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db := connectMongo()
	repo := repository.NewEventRepository(db)

	// Connect RabbitMQ
	rabbitConn := connectRabbitMQWithRetry()
	rabbitChannel := messaging.CreateChannel(rabbitConn)
	messaging.DeclareQueue(rabbitChannel, "ticket-reserved-queue")
	messaging.DeclareQueue(rabbitChannel, "payment-events-queue")

	// 🔁 Inizia ad ascoltare i PaymentEvent (questa era la riga mancante!)
	messaging.StartConsumingPaymentEvents(repo)

	srv := grpc.NewServer()
	ticket.RegisterTicketServiceServer(srv, &service.TicketService{
		Repo:          repo,
		RabbitChannel: rabbitChannel,
	})

	reflection.Register(srv)

	log.Println("Ticket Service is running on port 50052")
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
	"time"

	"github.com/rabbitmq/amqp091-go"

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
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatalf("❌ MongoDB ping error: %v", err)
	}

	log.Println("✅ Connessione a MongoDB (ticket-service) riuscita.")
	return client.Database("ticketdb")
}

func connectRabbitMQWithRetry() *amqp091.Connection {
	var conn *amqp091.Connection
	var err error
	for i := 1; i <= 5; i++ {
		conn, err = amqp091.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			log.Println("✅ Connessione a RabbitMQ riuscita")
			return conn
		}
		log.Printf("⚠️ Tentativo %d: RabbitMQ non disponibile. Riprovo in 3s...", i)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("❌ Impossibile connettersi a RabbitMQ dopo vari tentativi: %v", err)
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("❌ failed to listen: %v", err)
	}

	db := connectMongo()
	repo := repository.NewEventRepository(db)

	// Connessione a RabbitMQ con retry
	rabbitConn := connectRabbitMQWithRetry()
	rabbitChannel := messaging.CreateChannel(rabbitConn)

	// Dichiara le code
	messaging.DeclareQueue(rabbitChannel, "ticket-reserved-queue")
	messaging.DeclareQueue(rabbitChannel, "payment-events-queue")

	// Avvia il listener asincrono per gli eventi di pagamento
	messaging.StartConsumingPaymentEvents(repo)

	// Setup gRPC server
	srv := grpc.NewServer()
	ticket.RegisterTicketServiceServer(srv, &service.TicketService{
		Repo:          repo,
		RabbitChannel: rabbitChannel,
	})
	reflection.Register(srv)

	log.Println("🚀 Ticket Service is running on port 50052")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("❌ failed to serve: %v", err)
	}
}
*/

/*
package main

import (
	"context"
	"log"
	"net"

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
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatalf("❌ MongoDB ping error: %v", err)
	}

	log.Println("✅ Connessione a MongoDB (ticket-service) riuscita.")
	return client.Database("ticketdb")
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("❌ failed to listen: %v", err)
	}

	db := connectMongo()
	repo := repository.NewEventRepository(db)

	// ✅ Inizializza Circuit Breaker globale per RabbitMQ
	messaging.GlobalBreaker = messaging.NewCircuitBreaker("amqp://guest:guest@rabbitmq:5672/", []string{
		"ticket-reserved-queue",
		"payment-events-queue",
	})

	// 🚨 Verifica stato del breaker
	if messaging.GlobalBreaker == nil {
		log.Fatal("❌ GlobalBreaker non inizializzato")
	}
	if messaging.GlobalBreaker.Channel == nil {
		log.Fatal("❌ Channel è nil dopo init breaker → verifica che RabbitMQ sia up e accessibile!")
	}

	// 🔁 Listener eventi di pagamento (consumer con retry)
	messaging.StartConsumingPaymentEvents(repo)

	// 🛰️ Avvio server gRPC
	srv := grpc.NewServer()
	ticket.RegisterTicketServiceServer(srv, &service.TicketService{
		Repo: repo,
	})
	reflection.Register(srv)

	log.Println("🚀 Ticket Service is running on port 50052")
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
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatalf("❌ MongoDB ping error: %v", err)
	}

	log.Println("✅ Connessione a MongoDB (ticket-service) riuscita.")
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
			log.Println("✅ Circuit breaker inizializzato correttamente.")
			return
		}

		log.Printf("⏳ RabbitMQ non ancora pronto (tentativo %d/%d). Riprovo in 2s...", i, maxRetries)
		time.Sleep(2 * time.Second)
	}

	log.Fatal("❌ Impossibile inizializzare il Circuit Breaker: RabbitMQ non disponibile dopo vari tentativi.")
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("❌ failed to listen: %v", err)
	}

	db := connectMongo()
	repo := repository.NewEventRepository(db)

	// ⏳ Attendi RabbitMQ prima di inizializzare il breaker
	waitForRabbitAndInitBreaker()

	// 🔁 Listener eventi di pagamento (consumer con retry)
	messaging.StartConsumingPaymentEvents(repo)

	// 🛰️ Avvio server gRPC
	srv := grpc.NewServer()
	ticket.RegisterTicketServiceServer(srv, &service.TicketService{
		Repo: repo,
	})
	reflection.Register(srv)

	log.Println("🚀 Ticket Service is running on port 50052")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("❌ failed to serve: %v", err)
	}
}
