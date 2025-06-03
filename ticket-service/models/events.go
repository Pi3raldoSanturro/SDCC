package models

type TicketReservedEvent struct {
	EventId         string  `json:"eventId"`
	UserId          string  `json:"userId"`
	EventTicketId   string  `json:"eventTicketId"`
	Quantity        int32   `json:"quantity"`
	TotalAmount     float64 `json:"totalAmount"`
	EventInstanceId string  `json:"eventInstanceId"`
}

type PaymentSuccessEvent struct {
	EventId         string `json:"eventId"`
	UserId          string `json:"userId"`
	PaymentId       string `json:"paymentId"`
	EventInstanceId string `json:"eventInstanceId"`
}

type PaymentFailedEvent struct {
	EventId         string `json:"eventId"`
	UserId          string `json:"userId"`
	Reason          string `json:"reason"`
	EventInstanceId string `json:"eventInstanceId"`
}
