package domain

import "time"

// UserSummary represents lightweight user information for list endpoints.
type UserSummary struct {
	ID        string
	FullName  string
	Email     string
	Phone     string
	KYCStatus string
	RiskScore float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TransactionSummary represents lightweight transaction information.
type TransactionSummary struct {
	ID             string
	SenderUserID   string
	ReceiverUserID string
	Amount         float64
	Currency       string
	Type           string
	Status         string
	Channel        string
	Timestamp      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UserListResult captures paginated user list results.
type UserListResult struct {
	Items []UserSummary
	Total int64
}

// TransactionListResult captures paginated transaction list results.
type TransactionListResult struct {
	Items []TransactionSummary
	Total int64
}

