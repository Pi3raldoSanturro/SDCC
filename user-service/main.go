package main

import (
	"context"
	"log"
	"net"

	"user-service/proto/auth"
	"user-service/proto/user"
	"user-service/repository"
	"user-service/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func connectMongo() *mongo.Database {
	clientOptions := options.Client().ApplyURI("mongodb://mongo-user:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatalf("❌ MongoDB ping error: %v", err)
	}

	log.Println("✅ Connessione a MongoDB (user-service) riuscita.")
	return client.Database("userdb")
}

func connectAuthService() auth.AuthServiceClient {
	conn, err := grpc.Dial("auth-service:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Connessione a Auth-Service fallita: %v", err)
	}
	log.Println("✅ Connessione a Auth-Service riuscita.")
	return auth.NewAuthServiceClient(conn)
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ failed to listen: %v", err)
	}

	db := connectMongo()
	repo := repository.NewUserRepository(db)

	// 🔗 Client per Auth-Service
	authClient := connectAuthService()

	// ⛓️ Avvio gRPC Server
	srv := grpc.NewServer()
	user.RegisterUserServiceServer(srv, &service.UserService{
		Repo:       repo,
		AuthClient: authClient,
	})
	reflection.Register(srv)

	log.Println("🚀 User Service is running on port 50051")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("❌ failed to serve: %v", err)
	}
}
