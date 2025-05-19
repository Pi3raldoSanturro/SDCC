package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"payment-service/models"

	"go.mongodb.org/mongo-driver/mongo"
)

type PaymentRepository struct {
	collection *mongo.Collection
}

func NewPaymentRepository(db *mongo.Database) *PaymentRepository {
	return &PaymentRepository{
		collection: db.Collection("payments"),
	}
}

func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	_, err := r.collection.InsertOne(ctx, payment)
	return err
}

func (r *PaymentRepository) ExistsByEventInstanceID(ctx context.Context, eventInstanceId string) (bool, error) {
	filter := bson.M{"eventinstanceid": eventInstanceId}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
