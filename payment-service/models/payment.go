package models

type Payment struct {
	ID              string  `bson:"_id,omitempty"`
	UserID          string  `bson:"userId"`
	Amount          float64 `bson:"amount"`
	Success         bool    `bson:"success"`
	TransactionId   string  `bson:"transactionId"`
	EventInstanceId string  `bson:"eventInstanceId"`
}
