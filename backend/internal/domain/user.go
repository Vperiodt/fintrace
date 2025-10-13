package domain

import "time"

// Address captures structured address fields.
type Address struct {
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
}

// Attribute represents a normalized attribute used for shared links.
type Attribute struct {
	Type            string
	Value           string
	RawValue        string
	ConfidenceScore float64
}

// PaymentMethod represents a unique payment instrument associated with a user.
type PaymentMethod struct {
	ID          string
	MethodType  string
	Provider    string
	Masked      string
	Fingerprint string
	FirstUsedAt *time.Time
	LastUsedAt  *time.Time
}

// User aggregates the canonical user node data.
type User struct {
	ID             string
	FullName       string
	Email          string
	Phone          string
	Address        Address
	DateOfBirth    *time.Time
	KYCStatus      string
	RiskScore      float64
	Attributes     []Attribute
	PaymentMethods []PaymentMethod
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
