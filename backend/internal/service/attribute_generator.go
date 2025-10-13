package service

import (
	"strings"
	"time"

	"github.com/vanshika/fintrace/backend/internal/domain"
)

// DefaultAttributeGenerator implements AttributeGenerator using built-in normalization rules.
type DefaultAttributeGenerator struct{}

func (DefaultAttributeGenerator) FromUser(input UserInput) []domain.Attribute {
	var attrs []domain.Attribute

	if email := normalizeEmail(input.Email); email != "" {
		attrs = append(attrs, domain.Attribute{
			Type:            AttributeTypeEmail,
			Value:           hashValue(email),
			RawValue:        email,
			ConfidenceScore: defaultConfidenceScore,
		})
	}

	if phone := normalizePhone(input.Phone); phone != "" {
		attrs = append(attrs, domain.Attribute{
			Type:            AttributeTypePhone,
			Value:           hashValue(phone),
			RawValue:        phone,
			ConfidenceScore: defaultConfidenceScore,
		})
	}

	if addr := normalizeAddress(input.Address); strings.Trim(addr, "|") != "" {
		attrs = append(attrs, domain.Attribute{
			Type:            AttributeTypeAddress,
			Value:           hashValue(addr),
			RawValue:        addr,
			ConfidenceScore: 0.9,
		})
	}

	seenPaymentIdentifiers := make(map[string]struct{})
	for _, pm := range input.PaymentMethods {
		identifier := strings.TrimSpace(pm.Fingerprint)
		if identifier == "" {
			identifier = strings.TrimSpace(pm.ID)
		}
		if identifier == "" {
			continue
		}
		if _, exists := seenPaymentIdentifiers[identifier]; exists {
			continue
		}
		seenPaymentIdentifiers[identifier] = struct{}{}
		attrs = append(attrs, domain.Attribute{
			Type:            AttributeTypePayment,
			Value:           hashValue(identifier),
			RawValue:        identifier,
			ConfidenceScore: 0.95,
		})
	}

	return attrs
}

func (DefaultAttributeGenerator) FromTransaction(input TransactionInput) []domain.Attribute {
	var attrs []domain.Attribute

	if ip := strings.TrimSpace(input.IPAddress); ip != "" {
		attrs = append(attrs, domain.Attribute{
			Type:            AttributeTypeIPAddress,
			Value:           hashValue(ip),
			RawValue:        ip,
			ConfidenceScore: 0.85,
		})
	}

	if device := strings.TrimSpace(input.DeviceID); device != "" {
		attrs = append(attrs, domain.Attribute{
			Type:            AttributeTypeDevice,
			Value:           hashValue(device),
			RawValue:        device,
			ConfidenceScore: 0.9,
		})
	}

	if pm := strings.TrimSpace(input.PaymentMethodID); pm != "" {
		attrs = append(attrs, domain.Attribute{
			Type:            AttributeTypePayment,
			Value:           hashValue(pm),
			RawValue:        pm,
			ConfidenceScore: 0.9,
		})
	}

	// Ensure timestamp attribute can be used for clustering time-based analytics.
	attrs = append(attrs, domain.Attribute{
		Type:            "TX_DAY_BUCKET",
		Value:           hashValue(input.Timestamp.UTC().Format(time.DateOnly)),
		RawValue:        input.Timestamp.UTC().Format(time.RFC3339),
		ConfidenceScore: 0.5,
	})

	return attrs
}
