package main

import (
	"log"
	"net"

	"auth-service/proto"
	"auth-service/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	listener, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("❌ Impossibile avviare il listener: %v", err)
	}

	server := grpc.NewServer()
	proto.RegisterAuthServiceServer(server, &service.AuthServer{})

	reflection.Register(server)

	log.Println("🚀 Auth Service in ascolto sulla porta 50054")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("❌ Errore nell'avvio del server gRPC: %v", err)
	}
}
