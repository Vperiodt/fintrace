package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/vanshika/fintrace/backend/internal/domain"
	"github.com/vanshika/fintrace/backend/internal/repository"
)

// GraphRepository is the storage contract required by the relationship service.
type GraphRepository interface {
	UpsertUser(ctx context.Context, user domain.User) error
	UpsertTransaction(ctx context.Context, tx domain.Transaction, attributes []domain.Attribute) error
	FetchUserRelationships(ctx context.Context, userID string) (domain.UserRelationships, error)
	FetchTransactionRelationships(ctx context.Context, transactionID string) (domain.TransactionRelationships, error)
	ListUsers(ctx context.Context, opts repository.ListUsersOptions) (domain.UserListResult, error)
	ListTransactions(ctx context.Context, opts repository.ListTransactionsOptions) (domain.TransactionListResult, error)
	ShortestPathBetweenUsers(ctx context.Context, sourceID, targetID string) (domain.ShortestPath, error)
	ExportUsers(ctx context.Context) ([]domain.UserSummary, error)
	ExportTransactions(ctx context.Context) ([]domain.TransactionSummary, error)
}

// AttributeGenerator handles attribute extraction and hashing.
type AttributeGenerator interface {
	FromUser(input UserInput) []domain.Attribute
	FromTransaction(input TransactionInput) []domain.Attribute
}

// RelationshipService orchestrates ingestion logic and delegates persistence to the repository.
type RelationshipService struct {
	repo       GraphRepository
	attributes AttributeGenerator
	nowFn      func() time.Time
}

// PaginationMeta captures pagination metadata returned to API clients.
type PaginationMeta struct {
	Page       int
	PageSize   int
	TotalItems int64
	TotalPages int
}

// UsersPage represents paginated users with metadata.
type UsersPage struct {
	Items      []domain.UserSummary
	Pagination PaginationMeta
}

// TransactionsPage represents paginated transactions with metadata.
type TransactionsPage struct {
	Items      []domain.TransactionSummary
	Pagination PaginationMeta
}

// ListUsersParams defines filters for listing users.
type ListUsersParams struct {
	Page        int
	PageSize    int
	Search      string
	KYCStatus   string
	RiskMin     *float64
	RiskMax     *float64
	Country     string
	City        string
	EmailDomain string
	SortField   string
	SortOrder   string
}

// ListTransactionsParams defines filters for listing transactions.
type ListTransactionsParams struct {
	Page      int
	PageSize  int
	Search    string
	UserID    string
	Status    string
	Type      string
	MinAmount *float64
	MaxAmount *float64
	StartTime *time.Time
	EndTime   *time.Time
	Channel   string
	SortField string
	SortOrder string
}

// NewRelationshipService constructs a RelationshipService with optional overrides.
func NewRelationshipService(repo GraphRepository, gen AttributeGenerator) *RelationshipService {
	if gen == nil {
		gen = DefaultAttributeGenerator{}
	}
	return &RelationshipService{
		repo:       repo,
		attributes: gen,
		nowFn:      time.Now,
	}
}

// WithClock overrides the time provider (used primarily in tests).
func (s *RelationshipService) WithClock(nowFn func() time.Time) {
	if nowFn != nil {
		s.nowFn = nowFn
	}
}

// ListUsers retrieves paginated users matching provided filters.
func (s *RelationshipService) ListUsers(ctx context.Context, params ListUsersParams) (UsersPage, error) {
	page, pageSize := normalizePagination(params.Page, params.PageSize)
	offset := (page - 1) * pageSize

	riskMin := 0.0
	if params.RiskMin != nil {
		riskMin = clampFloat(*params.RiskMin, 0, 1)
	}
	riskMax := 0.0
	if params.RiskMax != nil {
		riskMax = clampFloat(*params.RiskMax, 0, 1)
		if riskMax > 0 && riskMax < riskMin {
			riskMax = riskMin
		}
	}

	result, err := s.repo.ListUsers(ctx, repository.ListUsersOptions{
		Offset:      offset,
		Limit:       pageSize,
		KYCStatus:   params.KYCStatus,
		RiskMin:     riskMin,
		RiskMax:     riskMax,
		Search:      params.Search,
		Country:     params.Country,
		City:        params.City,
		EmailDomain: params.EmailDomain,
		SortField:   params.SortField,
		SortOrder:   params.SortOrder,
	})
	if err != nil {
		return UsersPage{}, err
	}

	return UsersPage{
		Items:      result.Items,
		Pagination: buildPaginationMeta(page, pageSize, result.Total),
	}, nil
}

// ListTransactions retrieves paginated transactions matching filters.
func (s *RelationshipService) ListTransactions(ctx context.Context, params ListTransactionsParams) (TransactionsPage, error) {
	page, pageSize := normalizePagination(params.Page, params.PageSize)
	offset := (page - 1) * pageSize

	minAmount := 0.0
	if params.MinAmount != nil && *params.MinAmount > 0 {
		minAmount = *params.MinAmount
	}
	maxAmount := 0.0
	if params.MaxAmount != nil && *params.MaxAmount > 0 {
		maxAmount = *params.MaxAmount
		if maxAmount > 0 && maxAmount < minAmount {
			maxAmount = minAmount
		}
	}

	result, err := s.repo.ListTransactions(ctx, repository.ListTransactionsOptions{
		Offset:    offset,
		Limit:     pageSize,
		UserID:    params.UserID,
		Status:    params.Status,
		Type:      params.Type,
		MinAmount: minAmount,
		MaxAmount: maxAmount,
		Search:    params.Search,
		StartTs:   params.StartTime,
		EndTs:     params.EndTime,
		Channel:   params.Channel,
		SortField: params.SortField,
		SortOrder: params.SortOrder,
	})
	if err != nil {
		return TransactionsPage{}, err
	}

	return TransactionsPage{
		Items:      result.Items,
		Pagination: buildPaginationMeta(page, pageSize, result.Total),
	}, nil
}

// UpsertUser ingests a user payload, derives attributes, and persists graph mutations.
func (s *RelationshipService) UpsertUser(ctx context.Context, input UserInput) error {
	if input.ID == "" {
		return fmt.Errorf("user ID is required")
	}

	now := s.nowFn().UTC()
	createdAt := now
	updatedAt := now
	if input.CreatedAt != nil {
		createdAt = input.CreatedAt.UTC()
	}
	if input.UpdatedAt != nil {
		updatedAt = input.UpdatedAt.UTC()
	}

	user := domain.User{
		ID:          input.ID,
		FullName:    sanitizeString(input.FullName),
		Email:       normalizeEmail(input.Email),
		Phone:       normalizePhone(input.Phone),
		Address:     input.Address.ToDomainAddress(),
		DateOfBirth: input.DateOfBirth,
		KYCStatus:   input.KYCStatus,
		RiskScore:   input.RiskScore,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	paymentMethods := make([]domain.PaymentMethod, 0, len(input.PaymentMethods))
	for _, pm := range input.PaymentMethods {
		paymentMethods = append(paymentMethods, domain.PaymentMethod{
			ID:          pm.ID,
			MethodType:  pm.MethodType,
			Provider:    pm.Provider,
			Masked:      pm.Masked,
			Fingerprint: pm.Fingerprint,
			FirstUsedAt: pm.FirstUsedAt,
			LastUsedAt:  pm.LastUsedAt,
		})
	}
	user.PaymentMethods = paymentMethods

	attrs := s.attributes.FromUser(input)
	if len(input.Attributes) > 0 {
		attrs = append(attrs, convertCustomAttributes(input.Attributes)...)
	}
	user.Attributes = attrs

	return s.repo.UpsertUser(ctx, user)
}

// UpsertTransaction ingests a transaction payload, deriving edges and attributes before persisting.
func (s *RelationshipService) UpsertTransaction(ctx context.Context, input TransactionInput) error {
	if input.ID == "" {
		return fmt.Errorf("transaction ID is required")
	}
	if input.SenderUserID == "" || input.ReceiverUserID == "" {
		return fmt.Errorf("sender and receiver user IDs are required")
	}

	now := s.nowFn().UTC()
	createdAt := now
	updatedAt := now
	if input.CreatedAt != nil {
		createdAt = input.CreatedAt.UTC()
	}
	if input.UpdatedAt != nil {
		updatedAt = input.UpdatedAt.UTC()
	}

	tx := domain.Transaction{
		ID:              input.ID,
		SenderUserID:    input.SenderUserID,
		ReceiverUserID:  input.ReceiverUserID,
		Amount:          input.Amount,
		Currency:        input.Currency,
		Type:            input.Type,
		Status:          input.Status,
		Channel:         input.Channel,
		IPAddress:       input.IPAddress,
		DeviceID:        input.DeviceID,
		PaymentMethodID: input.PaymentMethodID,
		Timestamp:       input.Timestamp.UTC(),
		Metadata:        input.Metadata,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}

	attrs := s.attributes.FromTransaction(input)
	return s.repo.UpsertTransaction(ctx, tx, attrs)
}

// GetUserRelationships fetches relationship data for the provided user ID.
func (s *RelationshipService) GetUserRelationships(ctx context.Context, userID string) (domain.UserRelationships, error) {
	return s.repo.FetchUserRelationships(ctx, userID)
}

// GetTransactionRelationships fetches relationship data for the provided transaction ID.
func (s *RelationshipService) GetTransactionRelationships(ctx context.Context, txID string) (domain.TransactionRelationships, error) {
	return s.repo.FetchTransactionRelationships(ctx, txID)
}

func (s *RelationshipService) GetShortestPathBetweenUsers(ctx context.Context, sourceID, targetID string) (domain.ShortestPath, error) {
	sourceID = sanitizeString(sourceID)
	targetID = sanitizeString(targetID)
	if sourceID == "" || targetID == "" {
		return domain.ShortestPath{}, fmt.Errorf("sourceUserId and targetUserId are required")
	}
	return s.repo.ShortestPathBetweenUsers(ctx, sourceID, targetID)
}

func (s *RelationshipService) ExportUsers(ctx context.Context) ([]domain.UserSummary, error) {
	return s.repo.ExportUsers(ctx)
}

func (s *RelationshipService) ExportTransactions(ctx context.Context) ([]domain.TransactionSummary, error) {
	return s.repo.ExportTransactions(ctx)
}

func normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

func buildPaginationMeta(page, pageSize int, total int64) PaginationMeta {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
		if total > 0 && totalPages == 0 {
			totalPages = 1
		}
	}
	return PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if max > min && value > max {
		return max
	}
	return value
}

func convertCustomAttributes(attrs []AttributeInput) []domain.Attribute {
	var result []domain.Attribute
	for _, attr := range attrs {
		if attr.Type == "" || attr.Value == "" {
			continue
		}
		conf := attr.ConfidenceScore
		if conf == 0 {
			conf = defaultConfidenceScore
		}
		result = append(result, domain.Attribute{
			Type:            attr.Type,
			Value:           attr.Value,
			RawValue:        attr.RawValue,
			ConfidenceScore: conf,
		})
	}
	return result
}
