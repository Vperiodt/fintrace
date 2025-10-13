package service

import (
	"context"
	"errors"
	"sync"
)

// TaskError accumulates multiple errors produced during bulk ingestion.
type TaskError struct {
	Errors []error
}

func (e *TaskError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	msg := "multiple errors:"
	for _, err := range e.Errors {
		msg += " " + err.Error() + ";"
	}
	return msg
}

func (e *TaskError) append(err error) {
	if err == nil {
		return
	}
	e.Errors = append(e.Errors, err)
}

func (e *TaskError) asError() error {
	if len(e.Errors) == 0 {
		return nil
	}
	return e
}

// BulkIngestor processes large user and transaction datasets using worker pools.
type BulkIngestor struct {
	service *RelationshipService
	workers int
}

// NewBulkIngestor creates a new BulkIngestor instance with the provided concurrency.
func NewBulkIngestor(service *RelationshipService, workers int) *BulkIngestor {
	if workers <= 0 {
		workers = 4
	}
	return &BulkIngestor{
		service: service,
		workers: workers,
	}
}

// IngestUsers processes the provided user inputs concurrently.
func (bi *BulkIngestor) IngestUsers(ctx context.Context, users []UserInput) error {
	return bi.run(ctx, len(users), func(idx int) error {
		return bi.service.UpsertUser(ctx, users[idx])
	})
}

// IngestTransactions processes transaction inputs concurrently.
func (bi *BulkIngestor) IngestTransactions(ctx context.Context, txs []TransactionInput) error {
	return bi.run(ctx, len(txs), func(idx int) error {
		return bi.service.UpsertTransaction(ctx, txs[idx])
	})
}

func (bi *BulkIngestor) run(ctx context.Context, total int, workerFn func(idx int) error) error {
	if total == 0 {
		return nil
	}
	indexCh := make(chan int)
	errCh := make(chan error, total)
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for idx := range indexCh {
			if err := workerFn(idx); err != nil {
				select {
				case errCh <- err:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	for i := 0; i < bi.workers; i++ {
		wg.Add(1)
		go worker()
	}

Loop:
	for i := 0; i < total; i++ {
		select {
		case indexCh <- i:
		case <-ctx.Done():
			break Loop
		}
	}
	close(indexCh)
	wg.Wait()
	close(errCh)

	var taskErr TaskError
	for err := range errCh {
		if err == nil {
			continue
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		taskErr.append(err)
	}
	return taskErr.asError()
}
