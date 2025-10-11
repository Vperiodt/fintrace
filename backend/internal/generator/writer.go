package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteDataset serializes the dataset into users.json and transactions.json under the provided directory.
func WriteDataset(dataset Dataset, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	usersPath := filepath.Join(dir, "users.json")
	if err := writeJSON(usersPath, dataset.Users); err != nil {
		return err
	}

	transactionsPath := filepath.Join(dir, "transactions.json")
	if err := writeJSON(transactionsPath, dataset.Transactions); err != nil {
		return err
	}

	return nil
}

func writeJSON(path string, data any) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encode json for %s: %w", path, err)
	}
	return nil
}
