package testutil

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/db"
)

// TestStorageService wraps db for nats subscription.
type TestStorageService struct {
	Store  db.Db
}

// GetUserdataDb implements grassrootseconomics/visedriver/storage.StorageService.
func(ss *TestStorageService) GetUserdataDb(ctx context.Context) (db.Db, error) {
	return ss.Store, nil
}

// GetPersister implements grassrootseconomics/visedriver/storage.StorageService.
func(ts *TestStorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	return persist.NewPersister(ts.Store), nil
}

// GetResource implements grassrootseconomics/visedriver/storage.StorageService.
func(ts *TestStorageService) GetResource(ctx context.Context) (resource.Resource, error) {
	return nil, errors.New("not implemented")
}

// EnsureDbDir implements grassrootseconomics/visedriver/storage.StorageService.
func(ss *TestStorageService) EnsureDbDir() error {
	return nil
}
