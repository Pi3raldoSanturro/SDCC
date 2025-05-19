package service

import (
	"context"
	"fmt"
	"log"

	"auth-service/proto"
	"auth-service/token"
)

type AuthServer struct {
	proto.UnimplementedAuthServiceServer
}

func (s *AuthServer) GenerateToken(ctx context.Context, req *proto.AuthRequest) (*proto.AuthResponse, error) {
	jwtToken, err := token.GenerateJWT(req.UserId, req.Username, req.Role)
	if err != nil {
		log.Printf("❌ Errore nella generazione del token: %v", err)
		return nil, err
	}
	return &proto.AuthResponse{Token: jwtToken}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenValidationResponse, error) {
	claims, err := token.ValidateJWT(req.Token)
	if err != nil {
		return &proto.TokenValidationResponse{
			Valid:   false,
			Message: fmt.Sprintf("Token non valido: %v", err),
		}, nil
	}

	if claims == nil {
		return &proto.TokenValidationResponse{
			Valid:   false,
			Message: "Token non valido",
		}, nil
	}

	return &proto.TokenValidationResponse{
		Valid:    true,
		UserId:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		Message:  "Token valido",
	}, nil
}
