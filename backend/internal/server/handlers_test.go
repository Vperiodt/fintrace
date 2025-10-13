package server

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vanshika/fintrace/backend/internal/domain"
	"github.com/vanshika/fintrace/backend/internal/repository"
	"github.com/vanshika/fintrace/backend/internal/service"
)

type apiStubRepo struct {
	shortestPath           domain.ShortestPath
	usersList              domain.UserListResult
	txList                 domain.TransactionListResult
	userRelationships      domain.UserRelationships
	txRelationships        domain.TransactionRelationships
	exportUsersData        []domain.UserSummary
	exportTransactionsData []domain.TransactionSummary
}

func (a *apiStubRepo) UpsertUser(ctx context.Context, user domain.User) error { return nil }
func (a *apiStubRepo) UpsertTransaction(ctx context.Context, tx domain.Transaction, attrs []domain.Attribute) error {
	return nil
}
func (a *apiStubRepo) FetchUserRelationships(ctx context.Context, userID string) (domain.UserRelationships, error) {
	return a.userRelationships, nil
}
func (a *apiStubRepo) FetchTransactionRelationships(ctx context.Context, txID string) (domain.TransactionRelationships, error) {
	return a.txRelationships, nil
}
func (a *apiStubRepo) ListUsers(ctx context.Context, opts repository.ListUsersOptions) (domain.UserListResult, error) {
	return a.usersList, nil
}
func (a *apiStubRepo) ListTransactions(ctx context.Context, opts repository.ListTransactionsOptions) (domain.TransactionListResult, error) {
	return a.txList, nil
}
func (a *apiStubRepo) ShortestPathBetweenUsers(ctx context.Context, sourceID, targetID string) (domain.ShortestPath, error) {
	return a.shortestPath, nil
}
func (a *apiStubRepo) ExportUsers(ctx context.Context) ([]domain.UserSummary, error) {
	return a.exportUsersData, nil
}
func (a *apiStubRepo) ExportTransactions(ctx context.Context) ([]domain.TransactionSummary, error) {
	return a.exportTransactionsData, nil
}

func TestHandleShortestPath(t *testing.T) {
	repo := &apiStubRepo{
		shortestPath: domain.ShortestPath{
			SourceUserID: "USR-1",
			TargetUserID: "USR-2",
			Hops:         2,
			Nodes:        []domain.PathNode{{ID: "USR-1", Type: "User"}, {ID: "USR-2", Type: "User"}},
			Edges:        []domain.PathEdge{{Type: "SENT_TO", Source: "USR-1", Target: "USR-2"}},
		},
	}
	svc := service.NewRelationshipService(repo, nil)
	handlers := NewAPIHandlers(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	req := httptest.NewRequest(http.MethodGet, "/analytics/shortest-path?sourceUserId=USR-1&targetUserId=USR-2", nil)
	rec := httptest.NewRecorder()

	handlers.handleShortestPath(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload shortestPathResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Hops != 2 {
		t.Fatalf("expected hops 2, got %d", payload.Hops)
	}
	if len(payload.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(payload.Nodes))
	}
}

func TestHandleExportUsersCSV(t *testing.T) {
	repo := &apiStubRepo{
		exportUsersData: []domain.UserSummary{{ID: "USR-1", FullName: "Jane Doe", RiskScore: 0.4}},
	}
	svc := service.NewRelationshipService(repo, nil)
	handlers := NewAPIHandlers(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)

	req := httptest.NewRequest(http.MethodGet, "/export/users?format=csv", nil)
	rec := httptest.NewRecorder()

	handlers.handleExportUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Fatalf("expected text/csv content type, got %s", ct)
	}

	r := csv.NewReader(bytes.NewReader(rec.Body.Bytes()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse csv: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header + record), got %d", len(records))
	}
}
