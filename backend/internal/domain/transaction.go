package domain

import "time"

// Transaction models a transaction node in the graph.
type Transaction struct {
	ID              string
	SenderUserID    string
	ReceiverUserID  string
	Amount          float64
	Currency        string
	Type            string
	Status          string
	Channel         string
	IPAddress       string
	DeviceID        string
	PaymentMethodID string
	Timestamp       time.Time
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
