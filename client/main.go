package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	authpb "client/proto/auth-service"
	ticketpb "client/proto/ticket-service"
	userpb "client/proto/user-service"

	"google.golang.org/grpc"
)

const (
	maxRetries      = 3
	timeoutDuration = 5 * time.Second
)

func withRetry[T any](operation func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	var err error
	for i := 1; i <= maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		result, err := operation(ctx)
		if err == nil {
			return result, nil
		}
		fmt.Printf("Tentativo %d fallito: %v\n", i, err)
		time.Sleep(1 * time.Second)
	}
	return zero, fmt.Errorf("tutti i tentativi falliti: %v", err)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	userConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Connessione a user-service fallita: %v", err)
	}
	defer userConn.Close()
	userClient := userpb.NewUserServiceClient(userConn)

	ticketConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Connessione a ticket-service fallita: %v", err)
	}
	defer ticketConn.Close()
	ticketClient := ticketpb.NewTicketServiceClient(ticketConn)

	authConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Connessione a auth-service fallita: %v", err)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)

	fmt.Println("== Login o Registrazione ==")
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	loginResp, err := withRetry(func(ctx context.Context) (*userpb.LoginResponse, error) {
		return userClient.Login(ctx, &userpb.LoginRequest{
			Username: username,
			Password: password,
		})
	})

	if err != nil || !loginResp.Success {
		fmt.Println("Login fallito. Provo a registrare...")

		fmt.Print("Ruolo (admin/user): ")
		roleInput, _ := reader.ReadString('\n')
		roleInput = strings.TrimSpace(strings.ToLower(roleInput))
		if roleInput != "admin" {
			roleInput = "user"
		}

		_, err := withRetry(func(ctx context.Context) (*userpb.RegisterResponse, error) {
			return userClient.Register(ctx, &userpb.RegisterRequest{
				Username: username,
				Password: password,
				Role:     roleInput,
			})
		})
		if err != nil {
			log.Fatalf("Errore durante la registrazione: %v", err)
		}
		fmt.Println("Registrazione completata.")

		loginResp, err = withRetry(func(ctx context.Context) (*userpb.LoginResponse, error) {
			return userClient.Login(ctx, &userpb.LoginRequest{
				Username: username,
				Password: password,
			})
		})
		if err != nil {
			log.Fatalf("Errore di login dopo registrazione: %v", err)
		}
	}

	userID := loginResp.UserId
	role := loginResp.Role
	token := loginResp.Token

	fmt.Printf("✅ Autenticato come: %s (ruolo: %s)\n", username, role)

	for {
		fmt.Println("\n== Menu ==")
		fmt.Println("1. Visualizza eventi")
		fmt.Println("2. Acquista biglietto")
		if role == "admin" {
			fmt.Println("3. Aggiungi evento")
			fmt.Println("4. Rimuovi evento")
		}
		fmt.Println("5. Verifica validità token")
		fmt.Println("0. Esci")
		fmt.Print("Scelta: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			listResp, err := withRetry(func(ctx context.Context) (*ticketpb.ListEventsResponse, error) {
				return ticketClient.ListEvents(ctx, &ticketpb.ListEventsRequest{})
			})
			if err != nil {
				fmt.Printf("Errore nella richiesta eventi: %v\n", err)
				break
			}
			if listResp == nil || len(listResp.Events) == 0 {
				fmt.Println("Nessun evento disponibile.")
				break
			}
			fmt.Println("Eventi disponibili:")
			for _, e := range listResp.Events {
				fmt.Printf("- %s | Data: %s | Biglietti: %d | ID: %s\n", e.Name, e.Date, e.AvailableTickets, e.Id)
			}
		case "2":
			fmt.Print("ID evento: ")
			eventID, _ := reader.ReadString('\n')
			eventID = strings.TrimSpace(eventID)

			fmt.Print("Quantità: ")
			var qty int32
			fmt.Scanf("%d\n", &qty)

			resp, err := withRetry(func(ctx context.Context) (*ticketpb.PurchaseTicketResponse, error) {
				return ticketClient.PurchaseTicket(ctx, &ticketpb.PurchaseTicketRequest{
					EventId:  eventID,
					Quantity: qty,
					UserId:   userID,
				})
			})
			if err != nil {
				fmt.Printf("Errore nell'acquisto biglietto: %v\n", err)
			} else {
				fmt.Println(resp.Message)
			}
		case "3":
			if role != "admin" {
				fmt.Println("Accesso negato.")
				continue
			}
			fmt.Print("Nome evento: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			fmt.Print("Data (YYYY-MM-DD): ")
			date, _ := reader.ReadString('\n')
			date = strings.TrimSpace(date)

			fmt.Print("Biglietti disponibili: ")
			var qty int32
			fmt.Scanf("%d\n", &qty)

			resp, err := withRetry(func(ctx context.Context) (*ticketpb.AddEventResponse, error) {
				return ticketClient.AddEvent(ctx, &ticketpb.AddEventRequest{
					UserId:           userID,
					Name:             name,
					Date:             date,
					AvailableTickets: qty,
				})
			})
			if err != nil {
				fmt.Printf("Errore durante l'aggiunta evento: %v\n", err)
			} else {
				fmt.Println(resp.Message)
			}
		case "4":
			if role != "admin" {
				fmt.Println("Accesso negato.")
				continue
			}
			fmt.Print("ID evento da cancellare: ")
			eid, _ := reader.ReadString('\n')
			eid = strings.TrimSpace(eid)

			resp, err := withRetry(func(ctx context.Context) (*ticketpb.DeleteEventResponse, error) {
				return ticketClient.DeleteEvent(ctx, &ticketpb.DeleteEventRequest{
					UserId:  userID,
					EventId: eid,
				})
			})
			if err != nil {
				fmt.Printf("Errore durante la cancellazione: %v\n", err)
			} else {
				fmt.Println(resp.Message)
			}
		case "5":
			fmt.Println("Token corrente:", token)
			resp, err := withRetry(func(ctx context.Context) (*authpb.TokenValidationResponse, error) {
				return authClient.ValidateToken(ctx, &authpb.TokenRequest{Token: token})
			})
			if err != nil {
				fmt.Println("Errore durante validazione:", err)
			} else {
				fmt.Printf("✅ Token valido: %v, ruolo: %s, utente: %s\n", resp.Valid, resp.Role, resp.Username)
			}
		case "0":
			fmt.Println("Uscita...")
			return
		default:
			fmt.Println("Scelta non valida.")
		}
	}
}
