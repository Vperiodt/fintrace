package domain

import "time"

// DirectUserLink represents a user-to-user relationship through transactions.
type DirectUserLink struct {
	UserID        string
	LinkType      string
	Direction     string
	TransactionID string
	Amount        float64
	Currency      string
	Timestamp     *time.Time
}

// UserTransactionLink represents a user's participation in a transaction.
type UserTransactionLink struct {
	TransactionID string
	Role          string
	Amount        float64
	Currency      string
	Timestamp     *time.Time
}

// SharedAttributeLink groups users connected via a common attribute.
type SharedAttributeLink struct {
	AttributeType string
	AttributeHash string
	UserIDs       []string
}

// UserRelationships encapsulates all relationship views for a user.
type UserRelationships struct {
	UserID           string
	DirectLinks      []DirectUserLink
	Transactions     []UserTransactionLink
	SharedAttributes []SharedAttributeLink
}

// TransactionUserLink represents a user involved with a transaction.
type TransactionUserLink struct {
	UserID    string
	Role      string
	Amount    float64
	Currency  string
	Direction string
}

// LinkedTransaction captures attribute-based connections between transactions.
type LinkedTransaction struct {
	TransactionID string
	LinkType      string
	AttributeHash string
	Score         float64
	LastUpdated   *time.Time
}

// TransactionRelationships encapsulates relationship data for a transaction.
type TransactionRelationships struct {
	TransactionID      string
	Users              []TransactionUserLink
	LinkedTransactions []LinkedTransaction
}
