package service

import (
	"time"

	"github.com/vanshika/fintrace/backend/internal/domain"
)

// AddressInput mirrors domain.Address but keeps the separation between inbound payloads and storage models.
type AddressInput struct {
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
}

// PaymentMethodInput captures user provided payment instrument data.
type PaymentMethodInput struct {
	ID          string
	MethodType  string
	Provider    string
	Masked      string
	Fingerprint string
	FirstUsedAt *time.Time
	LastUsedAt  *time.Time
}

// AttributeInput allows callers to supply custom attributes beyond the default set.
type AttributeInput struct {
	Type            string
	Value           string
	RawValue        string
	ConfidenceScore float64
}

// UserInput is the inbound payload accepted by the relationship engine.
type UserInput struct {
	ID             string
	FullName       string
	Email          string
	Phone          string
	Address        AddressInput
	DateOfBirth    *time.Time
	KYCStatus      string
	RiskScore      float64
	PaymentMethods []PaymentMethodInput
	Attributes     []AttributeInput
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
}

// TransactionInput models data required to upsert a transaction and derive relationships.
type TransactionInput struct {
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
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

// ToDomainAddress converts the AddressInput to a domain.Address value.
func (in AddressInput) ToDomainAddress() domain.Address {
	return domain.Address{
		Line1:      in.Line1,
		Line2:      in.Line2,
		City:       in.City,
		State:      in.State,
		PostalCode: in.PostalCode,
		Country:    in.Country,
	}
}
