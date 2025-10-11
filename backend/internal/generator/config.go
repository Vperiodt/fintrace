package generator

// Config drives the synthetic data generator.
type Config struct {
	NumUsers                 int
	NumTransactions          int
	SharedAttributeChance    float64
	PaymentMethodShareChance float64
	IPShareChance            float64
	DeviceShareChance        float64
	Seed                     int64
}

// DefaultConfig returns baseline settings that satisfy the assignment requirements.
func DefaultConfig() Config {
	return Config{
		NumUsers:                 10000,
		NumTransactions:          100000,
		SharedAttributeChance:    0.35,
		PaymentMethodShareChance: 0.25,
		IPShareChance:            0.25,
		DeviceShareChance:        0.3,
		Seed:                     42,
	}
}
