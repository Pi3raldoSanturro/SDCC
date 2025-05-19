package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"ticket-service/messaging"
	"ticket-service/models"
	"ticket-service/proto"
	"ticket-service/repository"
)

type TicketService struct {
	Repo *repository.EventRepository
	ticket.UnimplementedTicketServiceServer
}

func (s *TicketService) Ping(ctx context.Context, req *ticket.PingRequest) (*ticket.PingResponse, error) {
	return &ticket.PingResponse{Message: "Ticket Service Pong!"}, nil
}

func (s *TicketService) ListEvents(ctx context.Context, req *ticket.ListEventsRequest) (*ticket.ListEventsResponse, error) {
	events, err := s.Repo.ListEvents(ctx)
	if err != nil {
		return nil, err
	}

	var protoEvents []*ticket.Event
	for _, e := range events {
		protoEvents = append(protoEvents, &ticket.Event{
			Id:               e.ID,
			Name:             e.Name,
			Date:             e.Date,
			AvailableTickets: e.AvailableTickets,
		})
	}

	return &ticket.ListEventsResponse{
		Events: protoEvents,
	}, nil
}

func (s *TicketService) AddEvent(ctx context.Context, req *ticket.AddEventRequest) (*ticket.AddEventResponse, error) {
	log.Println("⚙️ [AddEvent] Ricevuta richiesta per evento:", req.Name)

	if req.UserId == "" {
		log.Println("⚠️ [AddEvent] UserID mancante")
		return &ticket.AddEventResponse{
			Success: false,
			Message: "Missing user ID",
		}, nil
	}

	newEvent := &models.Event{
		Name:             req.Name,
		Date:             req.Date,
		AvailableTickets: req.AvailableTickets,
	}

	err := s.Repo.CreateEvent(ctx, newEvent)
	if err != nil {
		log.Printf("❌ [AddEvent] Errore inserimento MongoDB: %v", err)
		return nil, err
	}

	log.Println("✅ [AddEvent] Evento aggiunto con successo")
	return &ticket.AddEventResponse{
		Success: true,
		Message: "Evento aggiunto con successo",
	}, nil
}

func (s *TicketService) DeleteEvent(ctx context.Context, req *ticket.DeleteEventRequest) (*ticket.DeleteEventResponse, error) {
	log.Printf("⚙️ [DeleteEvent] Richiesta cancellazione per ID: %s", req.EventId)

	if req.UserId == "" || req.EventId == "" {
		return &ticket.DeleteEventResponse{
			Success: false,
			Message: "UserId e EventId sono obbligatori",
		}, nil
	}

	err := s.Repo.DeleteEventByID(ctx, req.EventId)
	if err != nil {
		log.Printf("❌ [DeleteEvent] Errore durante la cancellazione: %v", err)
		return &ticket.DeleteEventResponse{
			Success: false,
			Message: fmt.Sprintf("Errore durante la cancellazione: %v", err),
		}, nil
	}

	log.Println("✅ [DeleteEvent] Evento cancellato con successo")
	return &ticket.DeleteEventResponse{
		Success: true,
		Message: "Evento cancellato con successo",
	}, nil
}

/*
	func (s *TicketService) PurchaseTicket(ctx context.Context, req *ticket.PurchaseTicketRequest) (*ticket.PurchaseTicketResponse, error) {
		success, err := s.Repo.PurchaseTicket(ctx, req.EventId, req.Quantity)
		if err != nil {
			log.Printf("❌ Errore MongoDB durante la prenotazione: %v", err)
			return nil, err
		}
		if !success {
			return &ticket.PurchaseTicketResponse{
				Success: false,
				Message: "Biglietti non disponibili",
			}, nil
		}

		// 🆔 Genera ID una volta sola
		eventInstanceId := uuid.New().String()

		// Costruzione evento
		event := models.TicketReservedEvent{
			EventId:         req.EventId,
			UserId:          req.UserId,
			EventTicketId:   req.EventId,
			Quantity:        req.Quantity,
			TotalAmount:     float64(req.Quantity) * 21.0,
			EventInstanceId: eventInstanceId,
		}

		// Pubblica evento
		err = messaging.PublishMessage("ticket-reserved-queue", event)
		if err != nil {
			log.Printf("❌ Failed to publish TicketReservedEvent: %v", err)

			// 🔄 Rollback biglietti
			rollbackErr := s.Repo.RestoreTickets(ctx, req.EventId, req.Quantity)
			if rollbackErr != nil {
				log.Printf("⚠️ Errore durante il rollback dei biglietti: %v", rollbackErr)
			}

			return &ticket.PurchaseTicketResponse{
				Success: false,
				Message: "Errore di pubblicazione. Biglietti ripristinati.",
			}, nil
		}

		// Tutto OK
		return &ticket.PurchaseTicketResponse{
			Success: true,
			Message: "Ticket reserved successfully, waiting for payment...",
		}, nil
	}
*/
func (s *TicketService) PurchaseTicket(ctx context.Context, req *ticket.PurchaseTicketRequest) (*ticket.PurchaseTicketResponse, error) {
	// 1. Genera ID evento UNA sola volta
	eventInstanceId := uuid.New().String()

	event := models.TicketReservedEvent{
		EventId:         req.EventId,
		UserId:          req.UserId,
		EventTicketId:   req.EventId,
		Quantity:        req.Quantity,
		TotalAmount:     float64(req.Quantity) * 21.0,
		EventInstanceId: eventInstanceId,
	}

	// 2. Tenta il publish
	err := messaging.PublishMessage("ticket-reserved-queue", event)
	if err != nil {
		log.Printf("❌ Fallito publish TicketReservedEvent: %v", err)
		return &ticket.PurchaseTicketResponse{
			Success: false,
			Message: "Errore nel sistema, riprova più tardi.",
		}, nil
	}

	// 3. Solo se pubblicato → scala i biglietti
	success, err := s.Repo.PurchaseTicket(ctx, req.EventId, req.Quantity)
	if err != nil {
		log.Printf("❌ Errore MongoDB durante scalatura: %v", err)
		return nil, err
	}
	if !success {
		log.Printf("❌ Biglietti insufficienti per l'evento %s", req.EventId)
		return &ticket.PurchaseTicketResponse{
			Success: false,
			Message: "Biglietti esauriti",
		}, nil
	}

	// 4. Tutto OK
	return &ticket.PurchaseTicketResponse{
		Success: true,
		Message: "Biglietti prenotati, in attesa di pagamento...",
	}, nil
}
