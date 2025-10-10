package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/vanshika/fintrace/backend/internal/generator"
)

func main() {
	cfg := generator.DefaultConfig()
	var (
		users             = flag.Int("users", cfg.NumUsers, "number of users to generate")
		transactions      = flag.Int("transactions", cfg.NumTransactions, "number of transactions to generate")
		sharedChance      = flag.Float64("shared-attr-chance", cfg.SharedAttributeChance, "probability of reusing existing user attributes")
		pmShareChance     = flag.Float64("payment-share-chance", cfg.PaymentMethodShareChance, "probability of reusing existing payment methods")
		ipShareChance     = flag.Float64("ip-share-chance", cfg.IPShareChance, "probability of reusing existing IP addresses")
		deviceShareChance = flag.Float64("device-share-chance", cfg.DeviceShareChance, "probability of reusing existing device IDs")
		seed              = flag.Int64("seed", cfg.Seed, "random seed for deterministic generation")
		outputDir         = flag.String("output-dir", "data", "directory to write users.json and transactions.json")
		writeStdout       = flag.Bool("stdout", false, "write combined dataset to stdout instead of files")
	)
	flag.Parse()

	genCfg := generator.Config{
		NumUsers:                 *users,
		NumTransactions:          *transactions,
		SharedAttributeChance:    clampProbability(*sharedChance),
		PaymentMethodShareChance: clampProbability(*pmShareChance),
		IPShareChance:            clampProbability(*ipShareChance),
		DeviceShareChance:        clampProbability(*deviceShareChance),
		Seed:                     *seed,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	gen := generator.New(genCfg)
	dataset, err := gen.Generate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generation failed: %v\n", err)
		os.Exit(1)
	}

	if *writeStdout {
		if err := json.NewEncoder(os.Stdout).Encode(dataset); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write dataset to stdout: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := generator.WriteDataset(dataset, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write dataset: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Generated %d users and %d transactions into %s\n", len(dataset.Users), len(dataset.Transactions), *outputDir)
}

func clampProbability(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
