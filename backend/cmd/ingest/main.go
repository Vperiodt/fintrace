package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/vanshika/fintrace/backend/internal/config"
	"github.com/vanshika/fintrace/backend/internal/graph"
	"github.com/vanshika/fintrace/backend/internal/logging"
	"github.com/vanshika/fintrace/backend/internal/repository"
	"github.com/vanshika/fintrace/backend/internal/service"
)

var (
	errMissingDataset = errors.New("dataset not found")
)

func main() {
	var (
		datasetDir   = flag.String("dataset-dir", "./seed-data", "Directory containing users.json and transactions.json")
		usersPath    = flag.String("users", "", "Path to users.json (overrides dataset-dir)")
		transactions = flag.String("transactions", "", "Path to transactions.json (overrides dataset-dir)")
		workers      = flag.Int("workers", 4, "Number of concurrent workers for ingestion")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.Logging).With("component", "ingest")

	userFile, txFile, err := resolveDatasetPaths(*datasetDir, *usersPath, *transactions)
	if err != nil {
		logger.Error("dataset resolution failed", "error", err)
		os.Exit(1)
	}

	users, err := loadUserInputs(userFile)
	if err != nil {
		logger.Error("failed to load users", "error", err, "path", userFile)
		os.Exit(1)
	}
	if len(users) == 0 {
		logger.Error("users dataset empty", "path", userFile)
		os.Exit(1)
	}

	txs, err := loadTransactionInputs(txFile)
	if err != nil {
		logger.Error("failed to load transactions", "error", err, "path", txFile)
		os.Exit(1)
	}
	if len(txs) == 0 {
		logger.Error("transactions dataset empty", "path", txFile)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	graphClient, err := buildGraphClient(ctx, logger, cfg)
	if err != nil {
		logger.Error("failed to create graph client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := graphClient.Close(context.Background()); err != nil {
			logger.Warn("closing graph client failed", "error", err)
		}
	}()

	repo := repository.New(graphClient)
	svc := service.NewRelationshipService(repo, nil)
	ingestor := service.NewBulkIngestor(svc, *workers)

	start := time.Now()
	logger.Info("ingesting users", "count", len(users), "workers", *workers)
	if err := ingestor.IngestUsers(ctx, users); err != nil {
		logger.Error("user ingestion failed", "error", err)
		os.Exit(1)
	}

	logger.Info("ingesting transactions", "count", len(txs))
	if err := ingestor.IngestTransactions(ctx, txs); err != nil {
		logger.Error("transaction ingestion failed", "error", err)
		os.Exit(1)
	}

	logger.Info("ingestion complete", "duration", time.Since(start).String(), "users", len(users), "transactions", len(txs))
}

func resolveDatasetPaths(baseDir, usersPath, transactionsPath string) (string, string, error) {
	resolve := func(explicitPath, fallbackFile string) (string, error) {
		if explicitPath != "" {
			if _, err := os.Stat(explicitPath); err != nil {
				return "", fmt.Errorf("stat %s: %w", explicitPath, err)
			}
			return explicitPath, nil
		}
		path := filepath.Join(baseDir, fallbackFile)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("%w: %s", errMissingDataset, path)
		}
		return path, nil
	}

	usersFile, err := resolve(usersPath, "users.json")
	if err != nil {
		return "", "", err
	}
	txsFile, err := resolve(transactionsPath, "transactions.json")
	if err != nil {
		return "", "", err
	}
	return usersFile, txsFile, nil
}

func loadUserInputs(path string) ([]service.UserInput, error) {
	var users []service.UserInput
	if err := loadJSON(path, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func loadTransactionInputs(path string) ([]service.TransactionInput, error) {
	var txs []service.TransactionInput
	if err := loadJSON(path, &txs); err != nil {
		return nil, err
	}
	return txs, nil
}

func loadJSON(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func buildGraphClient(ctx context.Context, logger *slog.Logger, cfg config.Config) (graph.Client, error) {
	if cfg.Graph.URI == "" {
		return nil, fmt.Errorf("GRAPH_URI is required for ingestion")
	}
	opts := graph.Options{
		URI:            cfg.Graph.URI,
		Database:       cfg.Graph.Database,
		Username:       cfg.Graph.Username,
		Password:       cfg.Graph.Password,
		MaxConnections: cfg.Graph.MaxConnections,
	}
	client, err := graph.NewNeo4jClient(ctx, opts)
	if err != nil {
		return nil, err
	}
	if err := client.VerifyConnectivity(ctx); err != nil {
		_ = client.Close(ctx)
		return nil, err
	}
	logger.Info("connected to graph", "uri", cfg.Graph.URI, "database", cfg.Graph.Database)
	return client, nil
}
