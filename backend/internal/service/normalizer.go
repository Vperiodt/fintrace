package service

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	whitespaceRegex = regexp.MustCompile(`\s+`)
	nonDigitRegex   = regexp.MustCompile(`\D+`)
)

// AttributeType constants used by the relationship engine.
const (
	AttributeTypeEmail     = "EMAIL"
	AttributeTypePhone     = "PHONE"
	AttributeTypeAddress   = "ADDRESS"
	AttributeTypeIPAddress = "IP"
	AttributeTypeDevice    = "DEVICE"
	AttributeTypePayment   = "PAYMENT_METHOD"
	AttributeTypeBusiness  = "BUSINESS"
	AttributeTypeCustom    = "CUSTOM"
	defaultConfidenceScore = 1.0
)

// normalizeEmail lowercases and trims the provided email.
func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

// normalizePhone removes non-digit characters to produce a canonical representation.
func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = nonDigitRegex.ReplaceAllString(phone, "")
	if phone == "" {
		return ""
	}
	if strings.HasPrefix(phone, "00") {
		phone = phone[2:]
	}
	if !strings.HasPrefix(phone, "+") {
		// assume E.164 with missing plus
		phone = "+" + phone
	}
	return phone
}

// normalizeAddress concatenates address fields into a single canonical string.
func normalizeAddress(addr AddressInput) string {
	components := []string{
		strings.ToLower(strings.TrimSpace(addr.Line1)),
		strings.ToLower(strings.TrimSpace(addr.Line2)),
		strings.ToLower(strings.TrimSpace(addr.City)),
		strings.ToLower(strings.TrimSpace(addr.State)),
		strings.ToLower(strings.TrimSpace(addr.PostalCode)),
		strings.ToLower(strings.TrimSpace(addr.Country)),
	}
	return strings.Join(components, "|")
}

// hashValue returns a deterministic SHA-256 hash for the provided value.
func hashValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// sanitizeString collapses whitespace and trims the result.
func sanitizeString(value string) string {
	value = whitespaceRegex.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}
