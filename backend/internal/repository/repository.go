package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vanshika/fintrace/backend/internal/domain"
	"github.com/vanshika/fintrace/backend/internal/graph"
)

// ListUsersOptions defines filters and pagination for user listing.
type ListUsersOptions struct {
	Offset      int
	Limit       int
	KYCStatus   string
	RiskMin     float64
	RiskMax     float64
	Search      string
	Country     string
	City        string
	EmailDomain string
	SortField   string
	SortOrder   string
}

// ListTransactionsOptions defines filters and pagination for transaction listing.
type ListTransactionsOptions struct {
	Offset    int
	Limit     int
	UserID    string
	Status    string
	Type      string
	MinAmount float64
	MaxAmount float64
	Search    string
	StartTs   *time.Time
	EndTs     *time.Time
	Channel   string
	SortField string
	SortOrder string
}

// Repository encapsulates graph persistence operations.
type Repository struct {
	client graph.Client
}

// New instantiates a Repository backed by the supplied graph client.
func New(client graph.Client) *Repository {
	return &Repository{client: client}
}

// UpsertUser ensures a user node exists with the latest metadata and attribute edges.
func (r *Repository) UpsertUser(ctx context.Context, user domain.User) error {
	if user.ID == "" {
		return errors.New("user id is required")
	}

	params := map[string]any{
		"userId":         user.ID,
		"props":          userProperties(user),
		"attributes":     attributeParams(user.Attributes),
		"paymentMethods": paymentMethodParams(user.PaymentMethods),
	}

	_, err := r.client.ExecuteWrite(ctx, upsertUserCypher, params)
	if err != nil {
		return fmt.Errorf("upsert user %s: %w", user.ID, err)
	}
	return nil
}

// UpsertTransaction ensures a transaction node exists and all relationships are refreshed.
func (r *Repository) UpsertTransaction(ctx context.Context, tx domain.Transaction, attributes []domain.Attribute) error {
	if tx.ID == "" {
		return errors.New("transaction id is required")
	}
	if tx.SenderUserID == "" || tx.ReceiverUserID == "" {
		return errors.New("both sender and receiver user IDs are required")
	}

	params := map[string]any{
		"transactionId":   tx.ID,
		"senderId":        tx.SenderUserID,
		"receiverId":      tx.ReceiverUserID,
		"amount":          tx.Amount,
		"currency":        tx.Currency,
		"timestamp":       formatTime(tx.Timestamp),
		"props":           transactionProperties(tx),
		"attributes":      attributeParams(attributes),
		"paymentMethodId": tx.PaymentMethodID,
	}

	_, err := r.client.ExecuteWrite(ctx, upsertTransactionCypher, params)
	if err != nil {
		return fmt.Errorf("upsert transaction %s: %w", tx.ID, err)
	}

	return nil
}

// ListUsers returns paginated users matching provided filters.
func (r *Repository) ListUsers(ctx context.Context, opts ListUsersOptions) (domain.UserListResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	search := strings.ToLower(strings.TrimSpace(opts.Search))
	country := strings.ToLower(strings.TrimSpace(opts.Country))
	city := strings.ToLower(strings.TrimSpace(opts.City))
	emailDomain := strings.TrimSpace(opts.EmailDomain)
	if emailDomain != "" {
		emailDomain = strings.ToLower(strings.TrimPrefix(emailDomain, "@"))
		emailDomain = "@" + emailDomain
	}

	params := map[string]any{
		"kycStatus":   strings.ToUpper(strings.TrimSpace(opts.KYCStatus)),
		"riskMin":     opts.RiskMin,
		"riskMax":     opts.RiskMax,
		"search":      search,
		"country":     country,
		"city":        city,
		"emailDomain": emailDomain,
		"skip":        offset,
		"limit":       limit,
	}

	query := fmt.Sprintf(listUsersCypherTemplate, userFilterClause, userOrderClause(opts.SortField, opts.SortOrder))
	res, err := r.client.ExecuteRead(ctx, query, params)
	if err != nil {
		return domain.UserListResult{}, fmt.Errorf("list users query: %w", err)
	}

	var users []domain.UserSummary
	for _, record := range res.Records {
		item := domain.UserSummary{
			ID:        toString(record["userId"]),
			FullName:  toString(record["fullName"]),
			Email:     toString(record["email"]),
			Phone:     toString(record["phone"]),
			KYCStatus: toString(record["kycStatus"]),
			RiskScore: toFloat64(record["riskScore"]),
		}
		if created := toTimePtr(record["createdAt"]); created != nil {
			item.CreatedAt = *created
		}
		if updated := toTimePtr(record["updatedAt"]); updated != nil {
			item.UpdatedAt = *updated
		}
		users = append(users, item)
	}

	countQuery := fmt.Sprintf(countUsersCypherTemplate, userFilterClause)
	countRes, err := r.client.ExecuteRead(ctx, countQuery, params)
	if err != nil {
		return domain.UserListResult{}, fmt.Errorf("count users query: %w", err)
	}

	var total int64
	if len(countRes.Records) > 0 {
		switch v := countRes.Records[0]["total"].(type) {
		case int64:
			total = v
		case int:
			total = int64(v)
		case float64:
			total = int64(v)
		}
	}

	return domain.UserListResult{
		Items: users,
		Total: total,
	}, nil
}

// ListTransactions returns paginated transactions matching provided filters.
func (r *Repository) ListTransactions(ctx context.Context, opts ListTransactionsOptions) (domain.TransactionListResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	search := strings.ToLower(strings.TrimSpace(opts.Search))
	start := ""
	end := ""
	if opts.StartTs != nil && !opts.StartTs.IsZero() {
		start = opts.StartTs.UTC().Format(time.RFC3339)
	}
	if opts.EndTs != nil && !opts.EndTs.IsZero() {
		end = opts.EndTs.UTC().Format(time.RFC3339)
	}

	params := map[string]any{
		"userId":    strings.TrimSpace(opts.UserID),
		"status":    strings.ToUpper(strings.TrimSpace(opts.Status)),
		"type":      strings.ToUpper(strings.TrimSpace(opts.Type)),
		"minAmount": opts.MinAmount,
		"maxAmount": opts.MaxAmount,
		"search":    search,
		"skip":      offset,
		"limit":     limit,
		"startTs":   start,
		"endTs":     end,
		"channel":   strings.ToUpper(strings.TrimSpace(opts.Channel)),
	}

	orderClause := transactionOrderClause(opts.SortField, opts.SortOrder)
	query := fmt.Sprintf(listTransactionsCypherTemplate, transactionFilterClause, orderClause)
	res, err := r.client.ExecuteRead(ctx, query, params)
	if err != nil {
		return domain.TransactionListResult{}, fmt.Errorf("list transactions query: %w", err)
	}

	var txs []domain.TransactionSummary
	for _, record := range res.Records {
		item := domain.TransactionSummary{
			ID:             toString(record["transactionId"]),
			SenderUserID:   toString(record["senderId"]),
			ReceiverUserID: toString(record["receiverId"]),
			Amount:         toFloat64(record["amount"]),
			Currency:       toString(record["currency"]),
			Type:           toString(record["type"]),
			Status:         toString(record["status"]),
			Channel:        toString(record["channel"]),
		}
		if ts := toTimePtr(record["timestamp"]); ts != nil {
			item.Timestamp = *ts
		}
		if created := toTimePtr(record["createdAt"]); created != nil {
			item.CreatedAt = *created
		}
		if updated := toTimePtr(record["updatedAt"]); updated != nil {
			item.UpdatedAt = *updated
		}
		txs = append(txs, item)
	}

	countQuery := fmt.Sprintf(countTransactionsCypherTemplate, transactionFilterClause)
	countRes, err := r.client.ExecuteRead(ctx, countQuery, params)
	if err != nil {
		return domain.TransactionListResult{}, fmt.Errorf("count transactions query: %w", err)
	}

	var total int64
	if len(countRes.Records) > 0 {
		switch v := countRes.Records[0]["total"].(type) {
		case int64:
			total = v
		case int:
			total = int64(v)
		case float64:
			total = int64(v)
		}
	}

	return domain.TransactionListResult{
		Items: txs,
		Total: total,
	}, nil
}

// FetchUserRelationships returns a consolidated view of a user's relationships.
func (r *Repository) FetchUserRelationships(ctx context.Context, userID string) (domain.UserRelationships, error) {
	if userID == "" {
		return domain.UserRelationships{}, errors.New("user id is required")
	}

	relationships := domain.UserRelationships{
		UserID: userID,
	}

	if err := r.fetchUserDirectLinks(ctx, userID, &relationships); err != nil {
		return domain.UserRelationships{}, err
	}
	if err := r.fetchUserTransactions(ctx, userID, &relationships); err != nil {
		return domain.UserRelationships{}, err
	}
	if err := r.fetchUserSharedAttributes(ctx, userID, &relationships); err != nil {
		return domain.UserRelationships{}, err
	}

	return relationships, nil
}

// FetchTransactionRelationships returns relationships for a given transaction.
func (r *Repository) FetchTransactionRelationships(ctx context.Context, txID string) (domain.TransactionRelationships, error) {
	if txID == "" {
		return domain.TransactionRelationships{}, errors.New("transaction id is required")
	}

	result := domain.TransactionRelationships{
		TransactionID: txID,
	}

	if err := r.fetchTransactionUsers(ctx, txID, &result); err != nil {
		return domain.TransactionRelationships{}, err
	}
	if err := r.fetchLinkedTransactions(ctx, txID, &result); err != nil {
		return domain.TransactionRelationships{}, err
	}

	return result, nil
}

func (r *Repository) fetchUserDirectLinks(ctx context.Context, userID string, rel *domain.UserRelationships) error {
	res, err := r.client.ExecuteRead(ctx, userDirectLinksCypher, map[string]any{
		"userId": userID,
	})
	if err != nil {
		return fmt.Errorf("fetch user direct links: %w", err)
	}

	for _, record := range res.Records {
		link := domain.DirectUserLink{
			UserID:        toString(record["peerId"]),
			LinkType:      toString(record["linkType"]),
			Direction:     toString(record["direction"]),
			TransactionID: toString(record["transactionId"]),
			Amount:        toFloat64(record["amount"]),
			Currency:      toString(record["currency"]),
		}
		if ts := toTimePtr(record["timestamp"]); ts != nil {
			link.Timestamp = ts
		}
		rel.DirectLinks = append(rel.DirectLinks, link)
	}

	return nil
}

func (r *Repository) fetchUserTransactions(ctx context.Context, userID string, rel *domain.UserRelationships) error {
	res, err := r.client.ExecuteRead(ctx, userTransactionsCypher, map[string]any{
		"userId": userID,
	})
	if err != nil {
		return fmt.Errorf("fetch user transactions: %w", err)
	}

	for _, record := range res.Records {
		link := domain.UserTransactionLink{
			TransactionID: toString(record["transactionId"]),
			Role:          toString(record["role"]),
			Amount:        toFloat64(record["amount"]),
			Currency:      toString(record["currency"]),
		}
		if ts := toTimePtr(record["timestamp"]); ts != nil {
			link.Timestamp = ts
		}
		rel.Transactions = append(rel.Transactions, link)
	}
	return nil
}

func (r *Repository) fetchUserSharedAttributes(ctx context.Context, userID string, rel *domain.UserRelationships) error {
	res, err := r.client.ExecuteRead(ctx, userSharedAttributesCypher, map[string]any{
		"userId": userID,
	})
	if err != nil {
		return fmt.Errorf("fetch shared attributes: %w", err)
	}

	for _, record := range res.Records {
		usersRaw, ok := record["userIds"].([]any)
		if !ok {
			continue
		}
		var userIDs []string
		for _, u := range usersRaw {
			if s := toString(u); s != "" {
				userIDs = append(userIDs, s)
			}
		}
		rel.SharedAttributes = append(rel.SharedAttributes, domain.SharedAttributeLink{
			AttributeType: toString(record["attributeType"]),
			AttributeHash: toString(record["attributeHash"]),
			UserIDs:       userIDs,
		})
	}
	return nil
}

func (r *Repository) fetchTransactionUsers(ctx context.Context, txID string, rel *domain.TransactionRelationships) error {
	res, err := r.client.ExecuteRead(ctx, transactionUsersCypher, map[string]any{
		"transactionId": txID,
	})
	if err != nil {
		return fmt.Errorf("fetch transaction users: %w", err)
	}

	for _, record := range res.Records {
		rel.Users = append(rel.Users, domain.TransactionUserLink{
			UserID:    toString(record["userId"]),
			Role:      toString(record["role"]),
			Amount:    toFloat64(record["amount"]),
			Currency:  toString(record["currency"]),
			Direction: toString(record["direction"]),
		})
	}
	return nil
}

func (r *Repository) fetchLinkedTransactions(ctx context.Context, txID string, rel *domain.TransactionRelationships) error {
	res, err := r.client.ExecuteRead(ctx, transactionLinkedCypher, map[string]any{
		"transactionId": txID,
	})
	if err != nil {
		return fmt.Errorf("fetch linked transactions: %w", err)
	}

	for _, record := range res.Records {
		link := domain.LinkedTransaction{
			TransactionID: toString(record["otherTransactionId"]),
			LinkType:      toString(record["linkType"]),
			AttributeHash: toString(record["attributeHash"]),
			Score:         toFloat64(record["score"]),
		}
		if ts := toTimePtr(record["updatedAt"]); ts != nil {
			link.LastUpdated = ts
		}
		rel.LinkedTransactions = append(rel.LinkedTransactions, link)
	}
	return nil
}

// ShortestPathBetweenUsers finds the shortest path between two users if one exists.
func (r *Repository) ShortestPathBetweenUsers(ctx context.Context, sourceID, targetID string) (domain.ShortestPath, error) {
	if sourceID == "" || targetID == "" {
		return domain.ShortestPath{}, errors.New("source and target user IDs are required")
	}
	if sourceID == targetID {
		return domain.ShortestPath{
			SourceUserID: sourceID,
			TargetUserID: targetID,
			Nodes: []domain.PathNode{
				{ID: sourceID, Type: "User", Label: sourceID},
			},
			Hops: 0,
		}, nil
	}

	params := map[string]any{
		"sourceId": sourceID,
		"targetId": targetID,
	}

	res, err := r.client.ExecuteRead(ctx, shortestPathCypher, params)
	if err != nil {
		return domain.ShortestPath{}, fmt.Errorf("shortest path query: %w", err)
	}

	path := domain.ShortestPath{
		SourceUserID: sourceID,
		TargetUserID: targetID,
	}

	if len(res.Records) == 0 {
		return path, nil
	}

	record := res.Records[0]

	if nodesRaw, ok := record["nodes"].([]any); ok {
		for _, n := range nodesRaw {
			nodeMap, ok := n.(map[string]any)
			if !ok {
				continue
			}
			path.Nodes = append(path.Nodes, domain.PathNode{
				ID:     toString(nodeMap["id"]),
				Type:   toString(nodeMap["type"]),
				Label:  toString(nodeMap["label"]),
				Weight: toFloat64(nodeMap["weight"]),
			})
		}
	}

	if edgesRaw, ok := record["edges"].([]any); ok {
		for _, e := range edgesRaw {
			edgeMap, ok := e.(map[string]any)
			if !ok {
				continue
			}
			path.Edges = append(path.Edges, domain.PathEdge{
				Type:   toString(edgeMap["type"]),
				Source: toString(edgeMap["sourceId"]),
				Target: toString(edgeMap["targetId"]),
				Label:  toString(edgeMap["label"]),
				Weight: toFloat64(edgeMap["weight"]),
			})
		}
	}

	if hops, ok := record["hops"]; ok {
		switch v := hops.(type) {
		case int64:
			path.Hops = int(v)
		case int:
			path.Hops = v
		}
	}

	return path, nil
}

// ExportUsers returns all users for export purposes.
func (r *Repository) ExportUsers(ctx context.Context) ([]domain.UserSummary, error) {
	res, err := r.client.ExecuteRead(ctx, exportUsersCypher, nil)
	if err != nil {
		return nil, fmt.Errorf("export users query: %w", err)
	}
	users := make([]domain.UserSummary, 0, len(res.Records))
	for _, record := range res.Records {
		item := domain.UserSummary{
			ID:        toString(record["userId"]),
			FullName:  toString(record["fullName"]),
			Email:     toString(record["email"]),
			Phone:     toString(record["phone"]),
			KYCStatus: toString(record["kycStatus"]),
			RiskScore: toFloat64(record["riskScore"]),
		}
		if created := toTimePtr(record["createdAt"]); created != nil {
			item.CreatedAt = *created
		}
		if updated := toTimePtr(record["updatedAt"]); updated != nil {
			item.UpdatedAt = *updated
		}
		users = append(users, item)
	}
	return users, nil
}

// ExportTransactions returns all transactions for export purposes.
func (r *Repository) ExportTransactions(ctx context.Context) ([]domain.TransactionSummary, error) {
	res, err := r.client.ExecuteRead(ctx, exportTransactionsCypher, nil)
	if err != nil {
		return nil, fmt.Errorf("export transactions query: %w", err)
	}
	txs := make([]domain.TransactionSummary, 0, len(res.Records))
	for _, record := range res.Records {
		item := domain.TransactionSummary{
			ID:             toString(record["transactionId"]),
			SenderUserID:   toString(record["senderId"]),
			ReceiverUserID: toString(record["receiverId"]),
			Amount:         toFloat64(record["amount"]),
			Currency:       toString(record["currency"]),
			Type:           toString(record["type"]),
			Status:         toString(record["status"]),
			Channel:        toString(record["channel"]),
		}
		if ts := toTimePtr(record["timestamp"]); ts != nil {
			item.Timestamp = *ts
		}
		if created := toTimePtr(record["createdAt"]); created != nil {
			item.CreatedAt = *created
		}
		if updated := toTimePtr(record["updatedAt"]); updated != nil {
			item.UpdatedAt = *updated
		}
		txs = append(txs, item)
	}
	return txs, nil
}

func userProperties(u domain.User) map[string]any {
	props := map[string]any{
		"fullName":  u.FullName,
		"email":     u.Email,
		"phone":     u.Phone,
		"kycStatus": u.KYCStatus,
		"riskScore": u.RiskScore,
		"updatedAt": formatTime(u.UpdatedAt),
	}
	if !u.CreatedAt.IsZero() {
		props["createdAt"] = formatTime(u.CreatedAt)
	}

	if u.DateOfBirth != nil && !u.DateOfBirth.IsZero() {
		props["dateOfBirth"] = formatTime(*u.DateOfBirth)
	}

	props["addressLine1"] = u.Address.Line1
	props["addressLine2"] = u.Address.Line2
	props["addressCity"] = u.Address.City
	props["addressState"] = u.Address.State
	props["addressPostalCode"] = u.Address.PostalCode
	props["addressCountry"] = u.Address.Country

	return props
}

func paymentMethodParams(methods []domain.PaymentMethod) []map[string]any {
	result := make([]map[string]any, 0, len(methods))
	for _, pm := range methods {
		result = append(result, map[string]any{
			"id": pm.ID,
			"props": map[string]any{
				"methodType":  pm.MethodType,
				"provider":    pm.Provider,
				"masked":      pm.Masked,
				"fingerprint": pm.Fingerprint,
			},
			"firstUsedAt": formatTimePtr(pm.FirstUsedAt),
			"lastUsedAt":  formatTimePtr(pm.LastUsedAt),
		})
	}
	return result
}

func attributeParams(attrs []domain.Attribute) []map[string]any {
	result := make([]map[string]any, 0, len(attrs))
	for _, attr := range attrs {
		result = append(result, map[string]any{
			"type":       attr.Type,
			"value":      attr.Value,
			"rawValue":   attr.RawValue,
			"confidence": attr.ConfidenceScore,
			"score":      attr.ConfidenceScore,
		})
	}
	return result
}

func transactionProperties(tx domain.Transaction) map[string]any {
	props := map[string]any{
		"amount":          tx.Amount,
		"currency":        tx.Currency,
		"type":            tx.Type,
		"status":          tx.Status,
		"channel":         tx.Channel,
		"ipAddress":       tx.IPAddress,
		"deviceId":        tx.DeviceID,
		"paymentMethodId": tx.PaymentMethodID,
		"timestamp":       formatTime(tx.Timestamp),
		"updatedAt":       formatTime(tx.UpdatedAt),
	}

	if len(tx.Metadata) > 0 {
		if serialized, err := serializeMetadata(tx.Metadata); err == nil {
			props["metadataJson"] = serialized
		}
	}
	if !tx.CreatedAt.IsZero() {
		props["createdAt"] = formatTime(tx.CreatedAt)
	}
	return props
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func formatTimePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return formatTime(*t)
}

func serializeMetadata(metadata map[string]any) (string, error) {
	bytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func toString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func toFloat64(val any) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}

func toTimePtr(val any) *time.Time {
	switch v := val.(type) {
	case time.Time:
		return &v
	case string:
		if v == "" {
			return nil
		}
		if parsed, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return &parsed
		}
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			return &parsed
		}
	}
	return nil
}

const upsertUserCypher = `
MERGE (u:User {userId: $userId})
SET u += $props
WITH u
FOREACH (attr IN $attributes |
	MERGE (a:Attribute {attributeType: attr.type, value: attr.value})
	SET a.rawValue = attr.rawValue
	MERGE (u)-[ha:HAS_ATTRIBUTE]->(a)
	SET ha.confidenceScore = attr.confidence
)
FOREACH (pm IN $paymentMethods |
	MERGE (p:PaymentMethod {paymentMethodId: pm.id})
	SET p += pm.props
	MERGE (u)-[upm:USES_PAYMENT_METHOD]->(p)
	SET upm.firstUsedAt = pm.firstUsedAt
	SET upm.lastUsedAt = pm.lastUsedAt
)
RETURN u.userId AS userId
`

const upsertTransactionCypher = `
MATCH (sender:User {userId: $senderId})
MATCH (receiver:User {userId: $receiverId})
MERGE (t:Transaction {transactionId: $transactionId})
SET t += $props
MERGE (sender)-[ps:PARTICIPATED_IN {transactionId: $transactionId, role: "SENDER"}]->(t)
SET ps.amount = $amount,
	ps.currency = $currency,
	ps.timestamp = $timestamp
MERGE (receiver)-[pr:PARTICIPATED_IN {transactionId: $transactionId, role: "RECEIVER"}]->(t)
SET pr.amount = $amount,
	pr.currency = $currency,
	pr.timestamp = $timestamp
MERGE (sender)-[st:SENT_TO {transactionId: $transactionId}]->(receiver)
SET st.amount = $amount,
	st.currency = $currency,
	st.timestamp = $timestamp
MERGE (receiver)-[rt:RECEIVED_FROM {transactionId: $transactionId}]->(sender)
SET rt.amount = $amount,
	rt.currency = $currency,
	rt.timestamp = $timestamp
FOREACH (attr IN $attributes |
	MERGE (a:Attribute {attributeType: attr.type, value: attr.value})
	SET a.rawValue = attr.rawValue
	MERGE (t)-[hta:HAS_ATTRIBUTE]->(a)
	SET hta.origin = "TRANSACTION"
)
WITH t, $attributes AS attrs
UNWIND attrs AS attr
MATCH (a:Attribute {attributeType: attr.type, value: attr.value})
OPTIONAL MATCH (other:Transaction)-[:HAS_ATTRIBUTE]->(a)
WHERE other.transactionId <> $transactionId
WITH t, attr, collect(DISTINCT other) AS others
UNWIND others AS otherTx
MERGE (t)-[lt:LINKED_TO {attributeHash: attr.value, linkType: attr.type}]->(otherTx)
SET lt.score = attr.score,
    lt.updatedAt = datetime()
WITH DISTINCT t, $paymentMethodId AS pmId
OPTIONAL MATCH (pm:PaymentMethod {paymentMethodId: pmId})
FOREACH (_ IN CASE WHEN pmId = "" OR pm IS NULL THEN [] ELSE [1] END |
	MERGE (t)-[pmr:PAYMENT_METHOD_RELATES]->(pm)
	SET pmr.role = "SENDER"
)
RETURN t.transactionId AS transactionId
`

const listUsersCypherTemplate = `
MATCH (u:User)
%s
RETURN u.userId AS userId,
       u.fullName AS fullName,
       u.email AS email,
       u.phone AS phone,
       u.kycStatus AS kycStatus,
       u.riskScore AS riskScore,
       u.createdAt AS createdAt,
       u.updatedAt AS updatedAt
ORDER BY %s
SKIP $skip LIMIT $limit
`

const countUsersCypherTemplate = `
MATCH (u:User)
%s
RETURN count(u) AS total
`

const listTransactionsCypherTemplate = `
MATCH (t:Transaction)
%s
RETURN t.transactionId AS transactionId,
       t.amount AS amount,
       t.currency AS currency,
       t.type AS type,
       t.status AS status,
       t.channel AS channel,
       t.timestamp AS timestamp,
       t.createdAt AS createdAt,
       t.updatedAt AS updatedAt,
       head([(sender:User)-[:PARTICIPATED_IN {role: "SENDER"}]->(t) | sender.userId]) AS senderId,
       head([(receiver:User)-[:PARTICIPATED_IN {role: "RECEIVER"}]->(t) | receiver.userId]) AS receiverId
ORDER BY %s
SKIP $skip LIMIT $limit
`

const countTransactionsCypherTemplate = `
MATCH (t:Transaction)
%s
RETURN count(t) AS total
`

const userFilterClause = `
WHERE ($kycStatus = "" OR toUpper(u.kycStatus) = $kycStatus)
  AND ($riskMin <= 0 OR coalesce(u.riskScore, 0.0) >= $riskMin)
  AND ($riskMax <= 0 OR coalesce(u.riskScore, 0.0) <= $riskMax)
  AND ($search = "" OR toLower(u.fullName) CONTAINS $search OR toLower(u.email) CONTAINS $search OR toLower(u.userId) CONTAINS $search)
  AND ($country = "" OR toLower(coalesce(u.address.country, "")) = $country)
  AND ($city = "" OR toLower(coalesce(u.address.city, "")) = $city)
  AND ($emailDomain = "" OR toLower(u.email) ENDS WITH $emailDomain)
`

const transactionFilterClause = `
WHERE ($status = "" OR toUpper(t.status) = $status)
  AND ($type = "" OR toUpper(t.type) = $type)
  AND (
    $search = ""
    OR toLower(t.transactionId) CONTAINS $search
    OR EXISTS {
      MATCH (participant:User)-[:PARTICIPATED_IN]->(t)
      WHERE toLower(participant.userId) CONTAINS $search
        OR toLower(coalesce(participant.fullName, "")) CONTAINS $search
        OR toLower(coalesce(participant.email, "")) CONTAINS $search
    }
  )
  AND ($minAmount <= 0 OR coalesce(t.amount, 0.0) >= $minAmount)
  AND ($maxAmount <= 0 OR coalesce(t.amount, 0.0) <= $maxAmount)
  AND ($userId = "" OR EXISTS { MATCH (u:User {userId: $userId})-[:PARTICIPATED_IN]->(t) })
  AND ($startTs = "" OR t.timestamp >= datetime($startTs))
  AND ($endTs = "" OR t.timestamp <= datetime($endTs))
  AND ($channel = "" OR toUpper(t.channel) = $channel)
`

const shortestPathCypher = `
MATCH (source:User {userId: $sourceId}), (target:User {userId: $targetId})
MATCH path = shortestPath((source)-[:SENT_TO|RECEIVED_FROM|HAS_ATTRIBUTE|PARTICIPATED_IN|LINKED_TO|USES_PAYMENT_METHOD|PAYMENT_METHOD_RELATES*..6]-(target))
RETURN [n IN nodes(path) | {
  id: coalesce(n.userId, n.transactionId, n.paymentMethodId, n.value, toString(id(n))),
  label: coalesce(n.userId, n.transactionId, n.attributeType, n.paymentMethodId, n.value),
  type: head(labels(n)),
  weight: 1.0
}] AS nodes,
[rel IN relationships(path) | {
  type: type(rel),
  sourceId: coalesce(startNode(rel).userId, startNode(rel).transactionId, startNode(rel).paymentMethodId, startNode(rel).value, toString(id(startNode(rel)))),
  targetId: coalesce(endNode(rel).userId, endNode(rel).transactionId, endNode(rel).paymentMethodId, endNode(rel).value, toString(id(endNode(rel)))),
  label: type(rel),
  weight: 1.0
}] AS edges,
length(path) AS hops
`

const exportUsersCypher = `
MATCH (u:User)
RETURN u.userId AS userId,
       u.fullName AS fullName,
       u.email AS email,
       u.phone AS phone,
       u.kycStatus AS kycStatus,
       u.riskScore AS riskScore,
       u.createdAt AS createdAt,
       u.updatedAt AS updatedAt
ORDER BY u.userId
`

const exportTransactionsCypher = `
MATCH (t:Transaction)
RETURN t.transactionId AS transactionId,
       t.amount AS amount,
       t.currency AS currency,
       t.type AS type,
       t.status AS status,
       t.channel AS channel,
       t.timestamp AS timestamp,
       t.createdAt AS createdAt,
       t.updatedAt AS updatedAt,
       head([(sender:User)-[:PARTICIPATED_IN {role: "SENDER"}]->(t) | sender.userId]) AS senderId,
       head([(receiver:User)-[:PARTICIPATED_IN {role: "RECEIVER"}]->(t) | receiver.userId]) AS receiverId
ORDER BY datetime(t.timestamp) DESC
`

func userOrderClause(field, order string) string {
	dir := "ASC"
	if strings.EqualFold(order, "DESC") {
		dir = "DESC"
	}
	switch strings.ToLower(field) {
	case "fullname":
		return fmt.Sprintf("toLower(u.fullName) %s", dir)
	case "riskscore":
		return fmt.Sprintf("coalesce(u.riskScore, 0.0) %s", dir)
	case "createdat":
		return fmt.Sprintf("datetime(u.createdAt) %s", dir)
	case "updatedat":
		return fmt.Sprintf("datetime(u.updatedAt) %s", dir)
	default:
		return fmt.Sprintf("u.userId %s", dir)
	}
}

func transactionOrderClause(field, order string) string {
	dir := "DESC"
	if strings.EqualFold(order, "ASC") {
		dir = "ASC"
	}
	switch strings.ToLower(field) {
	case "amount":
		return fmt.Sprintf("coalesce(t.amount, 0.0) %s", dir)
	case "status":
		return fmt.Sprintf("toUpper(t.status) %s", dir)
	case "type":
		return fmt.Sprintf("toUpper(t.type) %s", dir)
	case "channel":
		return fmt.Sprintf("toUpper(t.channel) %s", dir)
	case "timestamp":
		return fmt.Sprintf("datetime(t.timestamp) %s", dir)
	case "createdat":
		return fmt.Sprintf("datetime(t.createdAt) %s", dir)
	case "updatedat":
		return fmt.Sprintf("datetime(t.updatedAt) %s", dir)
	case "transactionid":
		return fmt.Sprintf("t.transactionId %s", dir)
	default:
		return fmt.Sprintf("datetime(t.timestamp) %s", dir)
	}
}

const userDirectLinksCypher = `
MATCH (u:User {userId: $userId})-[r:SENT_TO|RECEIVED_FROM]->(peer:User)
RETURN peer.userId AS peerId,
       type(r) AS linkType,
       CASE WHEN type(r) = "SENT_TO" THEN "OUTBOUND" ELSE "INBOUND" END AS direction,
       r.transactionId AS transactionId,
       r.amount AS amount,
       r.currency AS currency,
       r.timestamp AS timestamp
`

const userTransactionsCypher = `
MATCH (u:User {userId: $userId})-[rel:PARTICIPATED_IN]->(t:Transaction)
RETURN t.transactionId AS transactionId,
       rel.role AS role,
       t.amount AS amount,
       t.currency AS currency,
       t.timestamp AS timestamp
`

const userSharedAttributesCypher = `
MATCH (u:User {userId: $userId})-[:HAS_ATTRIBUTE]->(a:Attribute)<-[:HAS_ATTRIBUTE]-(other:User)
WHERE other.userId <> $userId
RETURN a.attributeType AS attributeType,
       a.value AS attributeHash,
       collect(DISTINCT other.userId) AS userIds
`

const transactionUsersCypher = `
MATCH (t:Transaction {transactionId: $transactionId})<-[rel:PARTICIPATED_IN]-(user:User)
RETURN user.userId AS userId,
       rel.role AS role,
       rel.amount AS amount,
       rel.currency AS currency,
       CASE WHEN rel.role = "SENDER" THEN "OUTBOUND" ELSE "INBOUND" END AS direction
`

const transactionLinkedCypher = `
MATCH (t:Transaction {transactionId: $transactionId})-[link:LINKED_TO]->(other:Transaction)
RETURN other.transactionId AS otherTransactionId,
       link.linkType AS linkType,
       link.attributeHash AS attributeHash,
       link.score AS score,
       link.updatedAt AS updatedAt
`
