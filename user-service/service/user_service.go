package service

import (
	"context"
	"log"
	"user-service/models"
	"user-service/proto/user"
	"user-service/repository"

	authpb "user-service/proto/auth"
)

type UserService struct {
	Repo       *repository.UserRepository
	AuthClient authpb.AuthServiceClient
	user.UnimplementedUserServiceServer
}

func (s *UserService) Ping(ctx context.Context, req *user.PingRequest) (*user.PingResponse, error) {
	return &user.PingResponse{Message: "User Service Pong!"}, nil
}

func (s *UserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	exists, err := s.Repo.UserExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return &user.RegisterResponse{Message: "Username already exists"}, nil
	}

	role := req.Role
	if role != "admin" {
		role = "user"
	}

	newUser := &models.User{
		Username: req.Username,
		Password: req.Password,
		Role:     role,
	}

	err = s.Repo.CreateUser(ctx, newUser)
	newUserInDB, _ := s.Repo.FindByUsername(ctx, newUser.Username)
	newUser.ID = newUserInDB.ID
	if err != nil {
		return nil, err
	}

	// 🔐 Genera JWT con AuthService
	tokenResp, err := s.AuthClient.GenerateToken(ctx, &authpb.AuthRequest{
		UserId:   newUser.ID,
		Username: newUser.Username,
		Role:     newUser.Role,
	})
	if err != nil {
		log.Printf("Errore generazione token JWT in Register: %v", err)
		return &user.RegisterResponse{
			UserId:  newUser.ID,
			Message: "Registered but token generation failed",
		}, nil
	}

	return &user.RegisterResponse{
		UserId:  newUser.ID,
		Message: "Registration successful",
		Token:   tokenResp.Token,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	userFound, err := s.Repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return &user.LoginResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	if userFound.Password != req.Password {
		return &user.LoginResponse{
			Success: false,
			Message: "Incorrect password",
		}, nil
	}

	// 🔐 Chiede token a auth-service
	tokenResp, err := s.AuthClient.GenerateToken(ctx, &authpb.AuthRequest{
		UserId:   userFound.ID,
		Username: userFound.Username,
		Role:     userFound.Role,
	})
	if err != nil {
		log.Printf("Errore generazione token JWT in Login: %v", err)
		return &user.LoginResponse{
			UserId:  userFound.ID,
			Success: false,
			Message: "Login ok, ma token mancante",
			Role:    userFound.Role,
		}, nil
	}

	return &user.LoginResponse{
		UserId:  userFound.ID,
		Success: true,
		Message: "Login successful",
		Role:    userFound.Role,
		Token:   tokenResp.Token,
	}, nil
}
