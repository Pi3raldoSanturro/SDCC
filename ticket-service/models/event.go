package models

type Event struct {
	ID               string `bson:"_id,omitempty" json:"id,omitempty"`
	Name             string `bson:"name" json:"name"`
	Date             string `bson:"date" json:"date"`
	AvailableTickets int32  `bson:"availableTickets" json:"availableTickets"`
}
