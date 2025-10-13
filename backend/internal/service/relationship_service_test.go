package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vanshika/fintrace/backend/internal/domain"
	"github.com/vanshika/fintrace/backend/internal/repository"
)

type stubRepository struct {
	users               []domain.User
	transactions        []domain.Transaction
	transactionAttrs    [][]domain.Attribute
	transactionErr      error
	userErr             error
	usersList           domain.UserListResult
	transactionsList    domain.TransactionListResult
	listUsersErr        error
	listTransactionsErr error
	shortestPath        domain.ShortestPath
	shortestPathErr     error
	exportUsers         []domain.UserSummary
	exportTransactions  []domain.TransactionSummary
}

func (s *stubRepository) UpsertUser(ctx context.Context, user domain.User) error {
	if s.userErr != nil {
		return s.userErr
	}
	s.users = append(s.users, user)
	return nil
}

func (s *stubRepository) UpsertTransaction(ctx context.Context, tx domain.Transaction, attrs []domain.Attribute) error {
	if s.transactionErr != nil {
		return s.transactionErr
	}
	s.transactions = append(s.transactions, tx)
	s.transactionAttrs = append(s.transactionAttrs, attrs)
	return nil
}

func (s *stubRepository) FetchUserRelationships(ctx context.Context, userID string) (domain.UserRelationships, error) {
	return domain.UserRelationships{UserID: userID}, nil
}

func (s *stubRepository) FetchTransactionRelationships(ctx context.Context, txID string) (domain.TransactionRelationships, error) {
	return domain.TransactionRelationships{TransactionID: txID}, nil
}

func (s *stubRepository) ListUsers(ctx context.Context, opts repository.ListUsersOptions) (domain.UserListResult, error) {
	if s.listUsersErr != nil {
		return domain.UserListResult{}, s.listUsersErr
	}
	return s.usersList, nil
}

func (s *stubRepository) ListTransactions(ctx context.Context, opts repository.ListTransactionsOptions) (domain.TransactionListResult, error) {
	if s.listTransactionsErr != nil {
		return domain.TransactionListResult{}, s.listTransactionsErr
	}
	return s.transactionsList, nil
}

func (s *stubRepository) ShortestPathBetweenUsers(ctx context.Context, sourceID, targetID string) (domain.ShortestPath, error) {
	if s.shortestPathErr != nil {
		return domain.ShortestPath{}, s.shortestPathErr
	}
	return s.shortestPath, nil
}

func (s *stubRepository) ExportUsers(ctx context.Context) ([]domain.UserSummary, error) {
	if s.exportUsers != nil {
		return s.exportUsers, nil
	}
	return []domain.UserSummary{}, nil
}

func (s *stubRepository) ExportTransactions(ctx context.Context) ([]domain.TransactionSummary, error) {
	if s.exportTransactions != nil {
		return s.exportTransactions, nil
	}
	return []domain.TransactionSummary{}, nil
}

func TestRelationshipService_UpsertUser(t *testing.T) {
	repo := &stubRepository{}
	svc := NewRelationshipService(repo, nil)

	now := time.Date(2024, 4, 20, 12, 0, 0, 0, time.UTC)
	svc.WithClock(func() time.Time { return now })

	userInput := UserInput{
		ID:       "USR-1",
		FullName: "  Jane Doe ",
		Email:    "Jane.Doe@Example.com ",
		Phone:    " +1 (555) 123-4567 ",
		Address: AddressInput{
			Line1:      "123 Market St",
			City:       "San Francisco",
			State:      "CA",
			PostalCode: "94105",
			Country:    "US",
		},
		KYCStatus: "VERIFIED",
		RiskScore: 0.3,
		PaymentMethods: []PaymentMethodInput{
			{
				ID:          "PM-1",
				MethodType:  "CARD",
				Provider:    "VISA",
				Masked:      "4111********1111",
				Fingerprint: "FP-123",
			},
		},
	}

	if err := svc.UpsertUser(context.Background(), userInput); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repo.users) != 1 {
		t.Fatalf("expected 1 user persisted, got %d", len(repo.users))
	}

	user := repo.users[0]
	if user.Email != "jane.doe@example.com" {
		t.Errorf("expected normalized email, got %s", user.Email)
	}

	if len(user.Attributes) < 3 {
		t.Fatalf("expected at least 3 attributes (email, phone, address), got %d", len(user.Attributes))
	}
}

func TestRelationshipService_UpsertTransaction(t *testing.T) {
	repo := &stubRepository{}
	svc := NewRelationshipService(repo, nil)
	now := time.Date(2024, 4, 20, 12, 0, 0, 0, time.UTC)
	svc.WithClock(func() time.Time { return now })

	input := TransactionInput{
		ID:              "TX-1",
		SenderUserID:    "USR-1",
		ReceiverUserID:  "USR-2",
		Amount:          100,
		Currency:        "USD",
		Type:            "TRANSFER",
		Status:          "COMPLETED",
		Channel:         "WEB",
		IPAddress:       "203.0.113.24",
		DeviceID:        "device-abc",
		PaymentMethodID: "PM-1",
		Timestamp:       now,
	}

	if err := svc.UpsertTransaction(context.Background(), input); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repo.transactions) != 1 {
		t.Fatalf("expected 1 transaction persisted, got %d", len(repo.transactions))
	}

	if len(repo.transactionAttrs) == 0 {
		t.Fatalf("expected attributes recorded for transaction edges")
	}
}

func TestBulkIngestorAggregatesErrors(t *testing.T) {
	repo := &stubRepository{
		userErr: errors.New("boom"),
	}
	svc := NewRelationshipService(repo, nil)
	ingestor := NewBulkIngestor(svc, 2)

	err := ingestor.IngestUsers(context.Background(), []UserInput{
		{ID: "USR-1"},
		{ID: "USR-2"},
	})

	if err == nil {
		t.Fatalf("expected aggregated error, got nil")
	}
	taskErr, ok := err.(*TaskError)
	if !ok {
		t.Fatalf("expected TaskError type, got %T", err)
	}
	if len(taskErr.Errors) == 0 {
		t.Fatalf("expected TaskError to contain errors")
	}
}

func TestRelationshipService_ListUsers(t *testing.T) {
	repo := &stubRepository{
		usersList: domain.UserListResult{
			Items: []domain.UserSummary{
				{ID: "USR-1", FullName: "Jane Doe"},
			},
			Total: 1,
		},
	}
	svc := NewRelationshipService(repo, nil)

	result, err := svc.ListUsers(context.Background(), ListUsersParams{
		Page:     0,
		PageSize: 25,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Pagination.Page != 1 {
		t.Fatalf("expected page to default to 1, got %d", result.Pagination.Page)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 user, got %d", len(result.Items))
	}
}

func TestRelationshipService_ListTransactions(t *testing.T) {
	repo := &stubRepository{
		transactionsList: domain.TransactionListResult{
			Items: []domain.TransactionSummary{
				{ID: "TX-1", Amount: 100},
			},
			Total: 1,
		},
	}
	svc := NewRelationshipService(repo, nil)

	result, err := svc.ListTransactions(context.Background(), ListTransactionsParams{
		Page:     2,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Pagination.Page != 2 {
		t.Fatalf("expected page 2, got %d", result.Pagination.Page)
	}
	if result.Pagination.TotalPages == 0 {
		t.Fatalf("expected total pages > 0")
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(result.Items))
	}
}

func TestRelationshipService_GetShortestPathBetweenUsers(t *testing.T) {
	repo := &stubRepository{
		shortestPath: domain.ShortestPath{
			SourceUserID: "USR-1",
			TargetUserID: "USR-2",
			Hops:         3,
		},
	}
	svc := NewRelationshipService(repo, nil)
	path, err := svc.GetShortestPathBetweenUsers(context.Background(), "  USR-1 ", "USR-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path.Hops != 3 {
		t.Fatalf("expected hops 3, got %d", path.Hops)
	}
	if repo.shortestPath.SourceUserID != path.SourceUserID {
		t.Fatalf("expected source to match stub")
	}
}

func TestRelationshipService_ExportUsers(t *testing.T) {
	repo := &stubRepository{
		exportUsers: []domain.UserSummary{{ID: "USR-1"}},
	}
	svc := NewRelationshipService(repo, nil)
	users, err := svc.ExportUsers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
}

func TestRelationshipService_ExportTransactions(t *testing.T) {
	repo := &stubRepository{
		exportTransactions: []domain.TransactionSummary{{ID: "TX-1"}},
	}
	svc := NewRelationshipService(repo, nil)
	txs, err := svc.ExportTransactions(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs))
	}
}
