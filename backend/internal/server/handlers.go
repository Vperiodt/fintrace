package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vanshika/fintrace/backend/internal/service"
)

// APIHandlers exposes HTTP handlers for the REST API.
type APIHandlers struct {
	logger  *slog.Logger
	service *service.RelationshipService
}

// NewAPIHandlers constructs an APIHandlers instance.
func NewAPIHandlers(logger *slog.Logger, svc *service.RelationshipService) *APIHandlers {
	return &APIHandlers{
		logger:  logger,
		service: svc,
	}
}

func (h *APIHandlers) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createOrUpdateUser(w, r)
	case http.MethodGet:
		h.listUsers(w, r)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *APIHandlers) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createOrUpdateTransaction(w, r)
	case http.MethodGet:
		h.listTransactions(w, r)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *APIHandlers) handleUserRelationships(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/relationships/user/")
	userID = strings.Trim(userID, "/")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user ID is required")
		return
	}

	relationships, err := h.service.GetUserRelationships(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to fetch user relationships", "error", err, "userId", userID)
		writeError(w, http.StatusInternalServerError, "failed to fetch user relationships")
		return
	}

	response := userRelationshipsResponse{
		UserID:            userID,
		DirectConnections: []userDirectConnection{},
		Transactions:      []userTransactionLink{},
		SharedAttributes:  []sharedAttribute{},
	}

	for _, link := range relationships.DirectLinks {
		response.DirectConnections = append(response.DirectConnections, userDirectConnection{
			UserID:        link.UserID,
			LinkType:      link.LinkType,
			Direction:     link.Direction,
			TransactionID: link.TransactionID,
			Amount:        link.Amount,
			Currency:      link.Currency,
			Timestamp:     formatTimePtr(link.Timestamp),
		})
	}

	for _, tx := range relationships.Transactions {
		response.Transactions = append(response.Transactions, userTransactionLink{
			TransactionID: tx.TransactionID,
			Role:          tx.Role,
			Amount:        tx.Amount,
			Currency:      tx.Currency,
			Timestamp:     formatTimePtr(tx.Timestamp),
		})
	}

	for _, attr := range relationships.SharedAttributes {
		response.SharedAttributes = append(response.SharedAttributes, sharedAttribute{
			AttributeType: attr.AttributeType,
			AttributeHash: attr.AttributeHash,
			UserIDs:       attr.UserIDs,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) handleTransactionRelationships(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	txID := strings.TrimPrefix(r.URL.Path, "/relationships/transaction/")
	txID = strings.Trim(txID, "/")
	if txID == "" {
		writeError(w, http.StatusBadRequest, "transaction ID is required")
		return
	}

	relationships, err := h.service.GetTransactionRelationships(r.Context(), txID)
	if err != nil {
		h.logger.Error("failed to fetch transaction relationships", "error", err, "transactionId", txID)
		writeError(w, http.StatusInternalServerError, "failed to fetch transaction relationships")
		return
	}

	response := transactionRelationshipsResponse{
		TransactionID:      txID,
		Users:              []transactionUserLink{},
		LinkedTransactions: []linkedTransaction{},
	}
	for _, user := range relationships.Users {
		response.Users = append(response.Users, transactionUserLink{
			UserID:    user.UserID,
			Role:      user.Role,
			Amount:    user.Amount,
			Currency:  user.Currency,
			Direction: user.Direction,
		})
	}
	for _, link := range relationships.LinkedTransactions {
		response.LinkedTransactions = append(response.LinkedTransactions, linkedTransaction{
			TransactionID: link.TransactionID,
			LinkType:      link.LinkType,
			AttributeHash: link.AttributeHash,
			Score:         link.Score,
			UpdatedAt:     formatTimePtr(link.LastUpdated),
		})
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) createOrUpdateUser(w http.ResponseWriter, r *http.Request) {
	var payload userRequest
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if payload.UserID == "" {
		writeError(w, http.StatusBadRequest, "userId is required")
		return
	}

	input, err := payload.toServiceInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpsertUser(r.Context(), input); err != nil {
		h.logger.Error("failed to upsert user", "error", err, "userId", input.ID)
		writeError(w, http.StatusInternalServerError, "failed to persist user")
		return
	}

	respondJSON(w, http.StatusCreated, statusResponse{
		Status: "ok",
		ID:     input.ID,
	})
}

func (h *APIHandlers) listUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 50)
	search := query.Get("search")
	kycStatus := query.Get("kycStatus")
	country := query.Get("country")
	city := query.Get("city")
	emailDomain := query.Get("emailDomain")
	sortField := query.Get("sortField")
	sortOrder := query.Get("sortOrder")

	var riskMinPtr *float64
	if v := query.Get("riskMin"); v != "" {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid riskMin")
			return
		}
		riskMinPtr = &val
	}
	var riskMaxPtr *float64
	if v := query.Get("riskMax"); v != "" {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid riskMax")
			return
		}
		riskMaxPtr = &val
	}

	result, err := h.service.ListUsers(r.Context(), service.ListUsersParams{
		Page:        page,
		PageSize:    pageSize,
		Search:      search,
		KYCStatus:   kycStatus,
		RiskMin:     riskMinPtr,
		RiskMax:     riskMaxPtr,
		Country:     country,
		City:        city,
		EmailDomain: emailDomain,
		SortField:   sortField,
		SortOrder:   sortOrder,
	})
	if err != nil {
		h.logger.Error("failed to list users", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	resp := listUsersResponse{
		Pagination: paginationResponse{
			Page:       result.Pagination.Page,
			PageSize:   result.Pagination.PageSize,
			TotalItems: result.Pagination.TotalItems,
			TotalPages: result.Pagination.TotalPages,
		},
	}
	for _, item := range result.Items {
		resp.Items = append(resp.Items, userSummaryResponse{
			UserID:    item.ID,
			FullName:  item.FullName,
			Email:     item.Email,
			Phone:     item.Phone,
			KYCStatus: item.KYCStatus,
			RiskScore: item.RiskScore,
			CreatedAt: formatTime(item.CreatedAt),
			UpdatedAt: formatTime(item.UpdatedAt),
		})
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *APIHandlers) createOrUpdateTransaction(w http.ResponseWriter, r *http.Request) {
	var payload transactionRequest
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if payload.TransactionID == "" {
		writeError(w, http.StatusBadRequest, "transactionId is required")
		return
	}
	if payload.SenderUserID == "" || payload.ReceiverUserID == "" {
		writeError(w, http.StatusBadRequest, "senderUserId and receiverUserId are required")
		return
	}

	input, err := payload.toServiceInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpsertTransaction(r.Context(), input); err != nil {
		h.logger.Error("failed to upsert transaction", "error", err, "transactionId", input.ID)
		writeError(w, http.StatusInternalServerError, "failed to persist transaction")
		return
	}

	respondJSON(w, http.StatusCreated, statusResponse{
		Status: "ok",
		ID:     input.ID,
	})
}

func (h *APIHandlers) listTransactions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 50)
	search := query.Get("search")
	userID := query.Get("userId")
	status := query.Get("status")
	txType := query.Get("type")
	channel := query.Get("channel")
	sortField := query.Get("sortField")
	sortOrder := query.Get("sortOrder")

	var minAmountPtr *float64
	if v := query.Get("minAmount"); v != "" {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid minAmount")
			return
		}
		minAmountPtr = &val
	}
	var maxAmountPtr *float64
	if v := query.Get("maxAmount"); v != "" {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid maxAmount")
			return
		}
		maxAmountPtr = &val
	}

	var startPtr *time.Time
	if v := query.Get("start"); v != "" {
		ts, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start timestamp")
			return
		}
		startPtr = &ts
	}
	var endPtr *time.Time
	if v := query.Get("end"); v != "" {
		ts, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end timestamp")
			return
		}
		endPtr = &ts
	}

	result, err := h.service.ListTransactions(r.Context(), service.ListTransactionsParams{
		Page:      page,
		PageSize:  pageSize,
		Search:    search,
		UserID:    userID,
		Status:    status,
		Type:      txType,
		MinAmount: minAmountPtr,
		MaxAmount: maxAmountPtr,
		StartTime: startPtr,
		EndTime:   endPtr,
		Channel:   channel,
		SortField: sortField,
		SortOrder: sortOrder,
	})
	if err != nil {
		h.logger.Error("failed to list transactions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	resp := listTransactionsResponse{
		Pagination: paginationResponse{
			Page:       result.Pagination.Page,
			PageSize:   result.Pagination.PageSize,
			TotalItems: result.Pagination.TotalItems,
			TotalPages: result.Pagination.TotalPages,
		},
	}
	for _, item := range result.Items {
		resp.Items = append(resp.Items, transactionSummaryResponse{
			TransactionID:  item.ID,
			SenderUserID:   item.SenderUserID,
			ReceiverUserID: item.ReceiverUserID,
			Amount:         item.Amount,
			Currency:       item.Currency,
			Type:           item.Type,
			Status:         item.Status,
			Channel:        item.Channel,
			Timestamp:      formatTime(item.Timestamp),
			CreatedAt:      formatTime(item.CreatedAt),
			UpdatedAt:      formatTime(item.UpdatedAt),
		})
	}

	respondJSON(w, http.StatusOK, resp)
}

// --- Request & Response DTOs ---

type userRequest struct {
	UserID         string                 `json:"userId"`
	FullName       string                 `json:"fullName"`
	Email          string                 `json:"email"`
	Phone          string                 `json:"phone"`
	Address        addressRequest         `json:"address"`
	DateOfBirth    string                 `json:"dateOfBirth"`
	KYCStatus      string                 `json:"kycStatus"`
	RiskScore      float64                `json:"riskScore"`
	PaymentMethods []paymentMethodRequest `json:"paymentMethods"`
	Attributes     []attributeRequest     `json:"attributes"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
}

type addressRequest struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

type paymentMethodRequest struct {
	ID          string `json:"paymentMethodId"`
	MethodType  string `json:"methodType"`
	Provider    string `json:"provider"`
	Masked      string `json:"maskedNumber"`
	Fingerprint string `json:"fingerprint"`
	FirstUsedAt string `json:"firstUsedAt"`
	LastUsedAt  string `json:"lastUsedAt"`
}

type attributeRequest struct {
	Type            string  `json:"type"`
	Value           string  `json:"value"`
	RawValue        string  `json:"rawValue"`
	ConfidenceScore float64 `json:"confidenceScore"`
}

type transactionRequest struct {
	TransactionID   string         `json:"transactionId"`
	SenderUserID    string         `json:"senderUserId"`
	ReceiverUserID  string         `json:"receiverUserId"`
	Amount          float64        `json:"amount"`
	Currency        string         `json:"currency"`
	Type            string         `json:"type"`
	Status          string         `json:"status"`
	Channel         string         `json:"channel"`
	IPAddress       string         `json:"ipAddress"`
	DeviceID        string         `json:"deviceId"`
	PaymentMethodID string         `json:"paymentMethodId"`
	Timestamp       string         `json:"timestamp"`
	Metadata        map[string]any `json:"metadata"`
	CreatedAt       string         `json:"createdAt"`
	UpdatedAt       string         `json:"updatedAt"`
}

type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

type listUsersResponse struct {
	Items      []userSummaryResponse `json:"items"`
	Pagination paginationResponse    `json:"pagination"`
}

type listTransactionsResponse struct {
	Items      []transactionSummaryResponse `json:"items"`
	Pagination paginationResponse           `json:"pagination"`
}

type userSummaryResponse struct {
	UserID    string  `json:"userId"`
	FullName  string  `json:"fullName"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	KYCStatus string  `json:"kycStatus"`
	RiskScore float64 `json:"riskScore"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

type transactionSummaryResponse struct {
	TransactionID  string  `json:"transactionId"`
	SenderUserID   string  `json:"senderUserId"`
	ReceiverUserID string  `json:"receiverUserId"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Type           string  `json:"type"`
	Status         string  `json:"status"`
	Channel        string  `json:"channel"`
	Timestamp      string  `json:"timestamp"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type userRelationshipsResponse struct {
	UserID            string                 `json:"userId"`
	DirectConnections []userDirectConnection `json:"directConnections"`
	Transactions      []userTransactionLink  `json:"transactions"`
	SharedAttributes  []sharedAttribute      `json:"sharedAttributes"`
}

type userDirectConnection struct {
	UserID        string  `json:"userId"`
	LinkType      string  `json:"linkType"`
	Direction     string  `json:"direction"`
	TransactionID string  `json:"transactionId"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Timestamp     string  `json:"timestamp"`
}

type userTransactionLink struct {
	TransactionID string  `json:"transactionId"`
	Role          string  `json:"role"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Timestamp     string  `json:"timestamp"`
}

type sharedAttribute struct {
	AttributeType string   `json:"attributeType"`
	AttributeHash string   `json:"attributeHash"`
	UserIDs       []string `json:"connectedUsers"`
}

type transactionRelationshipsResponse struct {
	TransactionID      string                `json:"transactionId"`
	Users              []transactionUserLink `json:"users"`
	LinkedTransactions []linkedTransaction   `json:"linkedTransactions"`
}

type transactionUserLink struct {
	UserID    string  `json:"userId"`
	Role      string  `json:"role"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Direction string  `json:"direction"`
}

type linkedTransaction struct {
	TransactionID string  `json:"transactionId"`
	LinkType      string  `json:"linkType"`
	AttributeHash string  `json:"attributeHash"`
	Score         float64 `json:"score"`
	UpdatedAt     string  `json:"updatedAt"`
}

type statusResponse struct {
	Status string `json:"status"`
	ID     string `json:"id"`
}

// --- Helpers ---

func (req userRequest) toServiceInput() (service.UserInput, error) {
	var dobPtr *time.Time
	if req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return service.UserInput{}, fmtError("invalid dateOfBirth")
		}
		dobPtr = &dob
	}

	var createdPtr *time.Time
	if req.CreatedAt != "" {
		ts, err := time.Parse(time.RFC3339, req.CreatedAt)
		if err != nil {
			return service.UserInput{}, fmtError("invalid createdAt")
		}
		createdPtr = &ts
	}

	var updatedPtr *time.Time
	if req.UpdatedAt != "" {
		ts, err := time.Parse(time.RFC3339, req.UpdatedAt)
		if err != nil {
			return service.UserInput{}, fmtError("invalid updatedAt")
		}
		updatedPtr = &ts
	}

	paymentMethods := make([]service.PaymentMethodInput, 0, len(req.PaymentMethods))
	for _, pm := range req.PaymentMethods {
		pmInput := service.PaymentMethodInput{
			ID:          pm.ID,
			MethodType:  pm.MethodType,
			Provider:    pm.Provider,
			Masked:      pm.Masked,
			Fingerprint: pm.Fingerprint,
		}
		if pm.FirstUsedAt != "" {
			ts, err := time.Parse(time.RFC3339, pm.FirstUsedAt)
			if err != nil {
				return service.UserInput{}, fmtError("invalid firstUsedAt")
			}
			pmInput.FirstUsedAt = &ts
		}
		if pm.LastUsedAt != "" {
			ts, err := time.Parse(time.RFC3339, pm.LastUsedAt)
			if err != nil {
				return service.UserInput{}, fmtError("invalid lastUsedAt")
			}
			pmInput.LastUsedAt = &ts
		}
		paymentMethods = append(paymentMethods, pmInput)
	}

	attributes := make([]service.AttributeInput, 0, len(req.Attributes))
	for _, attr := range req.Attributes {
		attributes = append(attributes, service.AttributeInput{
			Type:            attr.Type,
			Value:           attr.Value,
			RawValue:        attr.RawValue,
			ConfidenceScore: attr.ConfidenceScore,
		})
	}

	return service.UserInput{
		ID:       req.UserID,
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Address: service.AddressInput{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			City:       req.Address.City,
			State:      req.Address.State,
			PostalCode: req.Address.PostalCode,
			Country:    req.Address.Country,
		},
		DateOfBirth:    dobPtr,
		KYCStatus:      req.KYCStatus,
		RiskScore:      req.RiskScore,
		PaymentMethods: paymentMethods,
		Attributes:     attributes,
		CreatedAt:      createdPtr,
		UpdatedAt:      updatedPtr,
	}, nil
}

func (req transactionRequest) toServiceInput() (service.TransactionInput, error) {
	if req.Timestamp == "" {
		return service.TransactionInput{}, fmtError("timestamp is required")
	}
	ts, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		return service.TransactionInput{}, fmtError("invalid timestamp")
	}

	var createdPtr *time.Time
	if req.CreatedAt != "" {
		ct, err := time.Parse(time.RFC3339, req.CreatedAt)
		if err != nil {
			return service.TransactionInput{}, fmtError("invalid createdAt")
		}
		createdPtr = &ct
	}

	var updatedPtr *time.Time
	if req.UpdatedAt != "" {
		ut, err := time.Parse(time.RFC3339, req.UpdatedAt)
		if err != nil {
			return service.TransactionInput{}, fmtError("invalid updatedAt")
		}
		updatedPtr = &ut
	}

	return service.TransactionInput{
		ID:              req.TransactionID,
		SenderUserID:    req.SenderUserID,
		ReceiverUserID:  req.ReceiverUserID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Type:            req.Type,
		Status:          req.Status,
		Channel:         req.Channel,
		IPAddress:       req.IPAddress,
		DeviceID:        req.DeviceID,
		PaymentMethodID: req.PaymentMethodID,
		Timestamp:       ts,
		Metadata:        req.Metadata,
		CreatedAt:       createdPtr,
		UpdatedAt:       updatedPtr,
	}, nil
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	return nil
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return fallback
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(ts *time.Time) string {
	if ts == nil || ts.IsZero() {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{
		"error": msg,
	})
}

func methodNotAllowed(w http.ResponseWriter, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func fmtError(msg string) error {
	return errors.New(msg)
}
