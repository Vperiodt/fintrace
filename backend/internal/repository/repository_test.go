package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vanshika/fintrace/backend/internal/domain"
	"github.com/vanshika/fintrace/backend/internal/graph"
)

func TestRepository_UpsertUser(t *testing.T) {
	mem := graph.NewMemoryClient()
	repo := New(mem)

	now := time.Now().UTC()
	dob := now.AddDate(-30, 0, 0)
	user := domain.User{
		ID:       "USR-001",
		FullName: "Jane Doe",
		Email:    "jane@example.com",
		Phone:    "+15551234567",
		Address: domain.Address{
			Line1:      "123 Market St",
			City:       "San Francisco",
			State:      "CA",
			PostalCode: "94105",
			Country:    "US",
		},
		DateOfBirth: &dob,
		KYCStatus:   "VERIFIED",
		RiskScore:   0.23,
		Attributes: []domain.Attribute{
			{Type: "EMAIL", Value: "hash:jane@example.com", RawValue: "jane@example.com", ConfidenceScore: 0.95},
		},
		PaymentMethods: []domain.PaymentMethod{
			{
				ID:          "PM-123",
				MethodType:  "CARD",
				Provider:    "VISA",
				Masked:      "4111********1111",
				Fingerprint: "fp-123",
				FirstUsedAt: &now,
				LastUsedAt:  &now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.UpsertUser(context.Background(), user); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	calls := mem.WriteCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 write query, got %d", len(calls))
	}

	call := calls[0]
	if call.Query != upsertUserCypher {
		t.Fatalf("unexpected query\nexpected:\n%s\ngot:\n%s", upsertUserCypher, call.Query)
	}

	if call.Params["userId"] != user.ID {
		t.Errorf("expected userId %s, got %v", user.ID, call.Params["userId"])
	}

	props, ok := call.Params["props"].(map[string]any)
	if !ok {
		t.Fatalf("expected props map, got %T", call.Params["props"])
	}

	if props["fullName"] != user.FullName {
		t.Errorf("fullName mismatch: want %s got %v", user.FullName, props["fullName"])
	}
	if props["kycStatus"] != user.KYCStatus {
		t.Errorf("kycStatus mismatch: want %s got %v", user.KYCStatus, props["kycStatus"])
	}

	attrParam, ok := call.Params["attributes"].([]map[string]any)
	if !ok || len(attrParam) != len(user.Attributes) {
		t.Fatalf("expected attributes slice of len %d got %T (len=%d)", len(user.Attributes), call.Params["attributes"], len(attrParam))
	}

	pmParam, ok := call.Params["paymentMethods"].([]map[string]any)
	if !ok || len(pmParam) != len(user.PaymentMethods) {
		t.Fatalf("expected payment methods slice of len %d got %T (len=%d)", len(user.PaymentMethods), call.Params["paymentMethods"], len(pmParam))
	}
}

func TestRepository_UpsertTransaction(t *testing.T) {
	mem := graph.NewMemoryClient()
	repo := New(mem)

	now := time.Now().UTC()
	tx := domain.Transaction{
		ID:              "TX-001",
		SenderUserID:    "USR-SENDER",
		ReceiverUserID:  "USR-RECEIVER",
		Amount:          150.75,
		Currency:        "USD",
		Type:            "TRANSFER",
		Status:          "COMPLETED",
		Channel:         "WEB",
		IPAddress:       "203.0.113.24",
		DeviceID:        "device-123",
		PaymentMethodID: "PM-123",
		Timestamp:       now,
		Metadata:        map[string]any{"merchantCategory": "REMITTANCE"},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	attrs := []domain.Attribute{
		{Type: "IP", Value: "hash:203.0.113.24", RawValue: "203.0.113.24", ConfidenceScore: 0.8},
	}

	if err := repo.UpsertTransaction(context.Background(), tx, attrs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	calls := mem.WriteCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 write query, got %d", len(calls))
	}

	call := calls[0]
	if call.Query != upsertTransactionCypher {
		t.Fatalf("unexpected transaction query\nexpected:\n%s\ngot:\n%s", upsertTransactionCypher, call.Query)
	}

	if call.Params["transactionId"] != tx.ID {
		t.Errorf("transactionId mismatch: want %s got %v", tx.ID, call.Params["transactionId"])
	}
	if call.Params["senderId"] != tx.SenderUserID {
		t.Errorf("senderId mismatch: want %s got %v", tx.SenderUserID, call.Params["senderId"])
	}
	if call.Params["receiverId"] != tx.ReceiverUserID {
		t.Errorf("receiverId mismatch: want %s got %v", tx.ReceiverUserID, call.Params["receiverId"])
	}

	attrParam, ok := call.Params["attributes"].([]map[string]any)
	if !ok || len(attrParam) != len(attrs) {
		t.Fatalf("expected attributes slice of len %d got %T (len=%d)", len(attrs), call.Params["attributes"], len(attrParam))
	}
}

func TestRepository_FetchUserRelationships(t *testing.T) {
	mem := graph.NewMemoryClient()
	repo := New(mem)

	ts := time.Now().UTC().Format(time.RFC3339Nano)

	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"peerId":        "USR-2",
			"linkType":      "SENT_TO",
			"direction":     "OUTBOUND",
			"transactionId": "TX-1",
			"amount":        200.0,
			"currency":      "USD",
			"timestamp":     ts,
		},
	}})
	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"transactionId": "TX-1",
			"role":          "SENDER",
			"amount":        200.0,
			"currency":      "USD",
			"timestamp":     ts,
		},
	}})
	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"attributeType": "EMAIL",
			"attributeHash": "hash:shared@example.com",
			"userIds":       []any{"USR-3", "USR-4"},
		},
	}})

	result, err := repo.FetchUserRelationships(context.Background(), "USR-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.DirectLinks) != 1 {
		t.Fatalf("expected 1 direct link, got %d", len(result.DirectLinks))
	}
	if result.DirectLinks[0].UserID != "USR-2" {
		t.Errorf("expected peer USR-2, got %s", result.DirectLinks[0].UserID)
	}

	if len(result.Transactions) != 1 && result.Transactions[0].TransactionID != "TX-1" {
		t.Fatalf("expected 1 transaction link with TX-1, got %+v", result.Transactions)
	}

	if len(result.SharedAttributes) != 1 {
		t.Fatalf("expected 1 shared attribute, got %d", len(result.SharedAttributes))
	}
	if len(result.SharedAttributes[0].UserIDs) != 2 {
		t.Errorf("expected 2 shared users, got %d", len(result.SharedAttributes[0].UserIDs))
	}
}

func TestRepository_FetchTransactionRelationships(t *testing.T) {
	mem := graph.NewMemoryClient()
	repo := New(mem)

	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"userId":    "USR-1",
			"role":      "SENDER",
			"amount":    150.00,
			"currency":  "USD",
			"direction": "OUTBOUND",
		},
	}})

	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"otherTransactionId": "TX-2",
			"linkType":           "IP",
			"attributeHash":      "hash:203.0.113.24",
			"score":              0.8,
			"updatedAt":          time.Now(),
		},
	}})

	result, err := repo.FetchTransactionRelationships(context.Background(), "TX-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Users) != 1 {
		t.Fatalf("expected 1 user link, got %d", len(result.Users))
	}

	if len(result.LinkedTransactions) != 1 {
		t.Fatalf("expected 1 linked transaction, got %d", len(result.LinkedTransactions))
	}

	if result.LinkedTransactions[0].TransactionID != "TX-2" {
		t.Errorf("expected linked transaction TX-2, got %s", result.LinkedTransactions[0].TransactionID)
	}
}

func TestRepository_ListUsers(t *testing.T) {
	mem := graph.NewMemoryClient()
	repo := New(mem)

	created := time.Now().UTC().Format(time.RFC3339Nano)
	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"userId":    "USR-1",
			"fullName":  "Jane Doe",
			"email":     "jane@example.com",
			"phone":     "+123",
			"kycStatus": "VERIFIED",
			"riskScore": 0.2,
			"createdAt": created,
			"updatedAt": created,
		},
	}})
	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"total": int64(1),
		},
	}})

	opts := ListUsersOptions{
		Limit:   10,
		Offset:  0,
		Search:  "jane",
		RiskMin: 0,
		RiskMax: 0,
	}

	result, err := repo.ListUsers(context.Background(), opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 user, got %d", len(result.Items))
	}

	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	calls := mem.ReadCalls()
	if len(calls) < 2 {
		t.Fatalf("expected at least 2 read queries, got %d", len(calls))
	}
	lastCalls := calls[len(calls)-2:]
	if !strings.Contains(lastCalls[0].Query, "ORDER BY u.userId ASC") {
		t.Errorf("unexpected ordering in list users query: %s", lastCalls[0].Query)
	}
}

func TestRepository_ListTransactions(t *testing.T) {
	mem := graph.NewMemoryClient()
	repo := New(mem)

	ts := time.Now().UTC().Format(time.RFC3339Nano)
	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"transactionId": "TX-1",
			"amount":        100.0,
			"currency":      "USD",
			"type":          "TRANSFER",
			"status":        "COMPLETED",
			"channel":       "WEB",
			"timestamp":     ts,
			"createdAt":     ts,
			"updatedAt":     ts,
			"senderId":      "USR-1",
			"receiverId":    "USR-2",
		},
	}})
	mem.PushReadResult(graph.Result{Records: []graph.Record{
		{
			"total": int64(1),
		},
	}})

	opts := ListTransactionsOptions{
		Limit:  10,
		Offset: 0,
		Search: "TX",
	}

	result, err := repo.ListTransactions(context.Background(), opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(result.Items))
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	calls := mem.ReadCalls()
	if len(calls) < 2 {
		t.Fatalf("expected at least 2 read queries, got %d", len(calls))
	}
	lastCalls := calls[len(calls)-2:]
	if !strings.Contains(lastCalls[0].Query, "ORDER BY datetime(t.timestamp) DESC") {
		t.Errorf("unexpected ordering in list transactions query: %s", lastCalls[0].Query)
	}
}
