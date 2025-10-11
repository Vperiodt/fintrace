package generator

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/vanshika/fintrace/backend/internal/service"
)

// Dataset contains the generated users and transactions.
type Dataset struct {
	Users        []service.UserInput        `json:"users"`
	Transactions []service.TransactionInput `json:"transactions"`
}

// Generator produces synthetic graph data aligned with the relationship engine schema.
type Generator struct {
	cfg           Config
	rand          *rand.Rand
	nameFragments nameFragments
	pools         attributePools
}

// New returns a configured Generator instance.
func New(cfg Config) *Generator {
	if cfg.NumUsers <= 0 {
		cfg.NumUsers = DefaultConfig().NumUsers
	}
	if cfg.NumTransactions <= 0 {
		cfg.NumTransactions = DefaultConfig().NumTransactions
	}
	if cfg.SharedAttributeChance <= 0 {
		cfg.SharedAttributeChance = DefaultConfig().SharedAttributeChance
	}
	if cfg.PaymentMethodShareChance <= 0 {
		cfg.PaymentMethodShareChance = DefaultConfig().PaymentMethodShareChance
	}
	if cfg.IPShareChance <= 0 {
		cfg.IPShareChance = DefaultConfig().IPShareChance
	}
	if cfg.DeviceShareChance <= 0 {
		cfg.DeviceShareChance = DefaultConfig().DeviceShareChance
	}
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}

	return &Generator{
		cfg:           cfg,
		rand:          rand.New(rand.NewSource(cfg.Seed)),
		nameFragments: defaultNameFragments(),
		pools:         attributePools{},
	}
}

// Generate synthesises users and transactions. It respects context cancellation.
func (g *Generator) Generate(ctx context.Context) (Dataset, error) {
	users := make([]service.UserInput, g.cfg.NumUsers)
	userPaymentMethods := make(map[string][]string, g.cfg.NumUsers)
	now := time.Now().UTC()

	for i := 0; i < g.cfg.NumUsers; i++ {
		if err := ctx.Err(); err != nil {
			return Dataset{}, err
		}

		userID := fmt.Sprintf("USR-%06d", i+1)
		createdAt := now.Add(-time.Duration(g.rand.Intn(365*24)) * time.Hour)
		updatedAt := createdAt.Add(time.Duration(g.rand.Intn(72)) * time.Hour)

		email := g.maybeSharedString(&g.pools.emails, g.cfg.SharedAttributeChance, func() string {
			return g.randomEmail(userID)
		})
		phone := g.maybeSharedString(&g.pools.phones, g.cfg.SharedAttributeChance, func() string {
			return g.randomPhone()
		})
		address := g.maybeSharedAddress()

		paymentMethods := g.maybePaymentMethods(userID)
		var pmIDs []string
		for _, pm := range paymentMethods {
			pmIDs = append(pmIDs, pm.ID)
		}
		userPaymentMethods[userID] = pmIDs

		dob := time.Date(1960+g.rand.Intn(30), time.Month(1+g.rand.Intn(12)), 1+g.rand.Intn(28), 0, 0, 0, 0, time.UTC)

		users[i] = service.UserInput{
			ID:             userID,
			FullName:       g.randomFullName(),
			Email:          email,
			Phone:          phone,
			Address:        address,
			DateOfBirth:    &dob,
			KYCStatus:      g.randomKYCStatus(),
			RiskScore:      g.rand.Float64(),
			PaymentMethods: paymentMethods,
			CreatedAt:      &createdAt,
			UpdatedAt:      &updatedAt,
		}
	}

	transactions := make([]service.TransactionInput, g.cfg.NumTransactions)
	merchantCategories := []string{"REMITTANCE", "PAYROLL", "E_COMMERCE", "CRYPTO", "GAMBLING", "DONATION"}

	for i := 0; i < g.cfg.NumTransactions; i++ {
		if err := ctx.Err(); err != nil {
			return Dataset{}, err
		}

		txID := fmt.Sprintf("TX-%07d", i+1)
		senderIdx := g.rand.Intn(len(users))
		receiverIdx := g.rand.Intn(len(users))
		if senderIdx == receiverIdx {
			receiverIdx = (receiverIdx + 1) % len(users)
		}

		sender := users[senderIdx]
		receiver := users[receiverIdx]
		amount := g.rand.Float64()*4900 + 100
		ip := g.maybeSharedString(&g.pools.ips, g.cfg.IPShareChance, g.randomIP)
		device := g.maybeSharedString(&g.pools.devices, g.cfg.DeviceShareChance, g.randomDeviceID)

		paymentIDs := userPaymentMethods[sender.ID]
		paymentMethodID := ""
		if len(paymentIDs) > 0 {
			paymentMethodID = paymentIDs[g.rand.Intn(len(paymentIDs))]
		}

		timestamp := now.Add(-time.Duration(g.rand.Intn(60*24)) * time.Minute)
		createdAt := timestamp.Add(-time.Duration(g.rand.Intn(120)) * time.Minute)
		updatedAt := timestamp.Add(time.Duration(g.rand.Intn(120)) * time.Minute)

		transactions[i] = service.TransactionInput{
			ID:              txID,
			SenderUserID:    sender.ID,
			ReceiverUserID:  receiver.ID,
			Amount:          amount,
			Currency:        "USD",
			Type:            g.randomTransactionType(),
			Status:          "COMPLETED",
			Channel:         g.randomChannel(),
			IPAddress:       ip,
			DeviceID:        device,
			PaymentMethodID: paymentMethodID,
			Timestamp:       timestamp,
			Metadata: map[string]any{
				"merchantCategory": merchantCategories[g.rand.Intn(len(merchantCategories))],
				"note":             g.randomNote(),
			},
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		}
	}

	return Dataset{Users: users, Transactions: transactions}, nil
}

type attributePools struct {
	emails    []string
	phones    []string
	addresses []service.AddressInput
	payments  []service.PaymentMethodInput
	ips       []string
	devices   []string
}

func (g *Generator) maybeSharedString(pool *[]string, chance float64, newValue func() string) string {
	if len(*pool) > 0 && g.rand.Float64() < chance {
		return (*pool)[g.rand.Intn(len(*pool))]
	}
	val := newValue()
	*pool = append(*pool, val)
	return val
}

func (g *Generator) maybeSharedAddress() service.AddressInput {
	if len(g.pools.addresses) > 0 && g.rand.Float64() < g.cfg.SharedAttributeChance {
		return g.pools.addresses[g.rand.Intn(len(g.pools.addresses))]
	}
	addr := service.AddressInput{
		Line1:      g.randomStreet(),
		City:       g.randomCity(),
		State:      g.randomState(),
		PostalCode: fmt.Sprintf("%05d", g.rand.Intn(99999)),
		Country:    "US",
	}
	g.pools.addresses = append(g.pools.addresses, addr)
	return addr
}

func (g *Generator) maybePaymentMethods(userID string) []service.PaymentMethodInput {
	count := 1 + g.rand.Intn(2)
	methods := make([]service.PaymentMethodInput, 0, count)
	for i := 0; i < count; i++ {
		if len(g.pools.payments) > 0 && g.rand.Float64() < g.cfg.PaymentMethodShareChance {
			pm := g.pools.payments[g.rand.Intn(len(g.pools.payments))]
			methods = append(methods, pm)
			continue
		}
		id := fmt.Sprintf("PM-%s-%d", userID, i+1)
		pm := service.PaymentMethodInput{
			ID:          id,
			MethodType:  g.randomMethodType(),
			Provider:    g.randomProvider(),
			Masked:      g.randomMaskedNumber(),
			Fingerprint: fmt.Sprintf("fp-%s-%d", userID, g.rand.Intn(1000)),
		}
		if g.rand.Float64() < 0.5 {
			g.pools.payments = append(g.pools.payments, pm)
		}
		methods = append(methods, pm)
	}
	return methods
}

func (g *Generator) randomFullName() string {
	return fmt.Sprintf("%s %s", g.nameFragments.first[g.rand.Intn(len(g.nameFragments.first))],
		g.nameFragments.last[g.rand.Intn(len(g.nameFragments.last))])
}

func (g *Generator) randomEmail(userID string) string {
	domain := g.nameFragments.domains[g.rand.Intn(len(g.nameFragments.domains))]
	return fmt.Sprintf("%s.%s@%s", g.nameFragments.first[g.rand.Intn(len(g.nameFragments.first))],
		g.nameFragments.last[g.rand.Intn(len(g.nameFragments.last))], domain)
}

func (g *Generator) randomPhone() string {
	return fmt.Sprintf("+1%03d%03d%04d", g.rand.Intn(900)+100, g.rand.Intn(900)+100, g.rand.Intn(10000))
}

func (g *Generator) randomStreet() string {
	return fmt.Sprintf("%d %s %s", g.rand.Intn(9999)+1,
		g.nameFragments.streetNames[g.rand.Intn(len(g.nameFragments.streetNames))],
		g.nameFragments.streetSuffix[g.rand.Intn(len(g.nameFragments.streetSuffix))])
}

func (g *Generator) randomCity() string {
	return g.nameFragments.cities[g.rand.Intn(len(g.nameFragments.cities))]
}

func (g *Generator) randomState() string {
	return g.nameFragments.states[g.rand.Intn(len(g.nameFragments.states))]
}

func (g *Generator) randomKYCStatus() string {
	options := []string{"PENDING", "VERIFIED", "REVIEW"}
	return options[g.rand.Intn(len(options))]
}

func (g *Generator) randomMethodType() string {
	types := []string{"CARD", "BANK_ACCOUNT", "WALLET"}
	return types[g.rand.Intn(len(types))]
}

func (g *Generator) randomProvider() string {
	providers := []string{"VISA", "MASTERCARD", "AMEX", "DISCOVER", "PAYPAL", "STRIPE"}
	return providers[g.rand.Intn(len(providers))]
}

func (g *Generator) randomMaskedNumber() string {
	return fmt.Sprintf("%04d********%04d", g.rand.Intn(10000), g.rand.Intn(10000))
}

func (g *Generator) randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", g.rand.Intn(223)+1, g.rand.Intn(256), g.rand.Intn(256), g.rand.Intn(256))
}

func (g *Generator) randomDeviceID() string {
	return fmt.Sprintf("device-%06d", g.rand.Intn(999999))
}

func (g *Generator) randomTransactionType() string {
	types := []string{"TRANSFER", "PAYMENT", "WITHDRAWAL", "DEPOSIT"}
	return types[g.rand.Intn(len(types))]
}

func (g *Generator) randomChannel() string {
	channels := []string{"WEB", "MOBILE", "POS", "API"}
	return channels[g.rand.Intn(len(channels))]
}

func (g *Generator) randomNote() string {
	notes := []string{"Invoice settlement", "Freelance payout", "Peer transfer", "Market purchase", "Crypto off-ramp"}
	return notes[g.rand.Intn(len(notes))]
}

type nameFragments struct {
	first        []string
	last         []string
	domains      []string
	streetNames  []string
	streetSuffix []string
	cities       []string
	states       []string
}

func defaultNameFragments() nameFragments {
	return nameFragments{
		first:        []string{"Jane", "John", "Alex", "Priya", "Liu", "Maria", "Omar", "Sofia", "Noah", "Emma", "Lucas", "Mia", "Ava", "Ethan", "Zara"},
		last:         []string{"Doe", "Smith", "Chen", "Patel", "Garcia", "Khan", "Kim", "Ivanov", "Nguyen", "Silva", "Brown", "Lee"},
		domains:      []string{"example.com", "mail.com", "fintrace.io", "payments.net", "securepay.org"},
		streetNames:  []string{"Market", "Mission", "Broadway", "Fifth", "Sunset", "Park", "Cedar", "Oak", "Pine", "Ash"},
		streetSuffix: []string{"St", "Ave", "Blvd", "Ln", "Rd", "Way"},
		cities:       []string{"San Francisco", "New York", "Seattle", "Austin", "Chicago", "Miami", "Denver", "Boston", "Los Angeles"},
		states:       []string{"CA", "NY", "WA", "TX", "IL", "FL", "CO", "MA"},
	}
}
