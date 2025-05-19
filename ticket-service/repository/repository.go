package repository

import (
	"context"
	"errors"
	"ticket-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventRepository struct {
	collection    *mongo.Collection
	processedColl *mongo.Collection
}

func NewEventRepository(db *mongo.Database) *EventRepository {
	return &EventRepository{
		collection:    db.Collection("events"),
		processedColl: db.Collection("processed_events"), // ✅ nuova collezione
	}
}

func (r *EventRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *EventRepository) ListEvents(ctx context.Context) ([]*models.Event, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*models.Event
	for cursor.Next(ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, nil
}

func (r *EventRepository) PurchaseTicket(ctx context.Context, eventId string, quantity int32) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(eventId)
	if err != nil {
		return false, err
	}

	update := bson.M{
		"$inc": bson.M{"availableTickets": -quantity},
	}
	filter := bson.M{
		"_id":              objID,
		"availableTickets": bson.M{"$gte": quantity},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	if result.MatchedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (r *EventRepository) MarkTicketAsPaid(ctx context.Context, eventId string) error {
	objID, err := primitive.ObjectIDFromHex(eventId)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": "paid"}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *EventRepository) RestoreTickets(ctx context.Context, eventId string, quantity int32) error {
	objID, err := primitive.ObjectIDFromHex(eventId)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$inc": bson.M{"availableTickets": quantity}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *EventRepository) DeleteEventByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("ID non valido per ObjectID")
	}

	filter := bson.M{"_id": objectID}
	_, err = r.collection.DeleteOne(ctx, filter)
	return err
}

//
// ✅ DE-DUPLICAZIONE EVENTI PAGAMENTO
//

func (r *EventRepository) HasProcessedEvent(ctx context.Context, eventInstanceID string) (bool, error) {
	filter := bson.M{"eventInstanceId": eventInstanceID}
	count, err := r.processedColl.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *EventRepository) MarkEventAsProcessed(ctx context.Context, eventInstanceID string) error {
	_, err := r.processedColl.InsertOne(ctx, bson.M{"eventInstanceId": eventInstanceID})
	return err
}
