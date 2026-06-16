package stores

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/shared/atomicfile"
	"github.com/jimjibone/wh/v1/shared/paths"
)

// Store holds key value pairs.
type Store interface {
	// Set a key in the store.
	Set(key string, value []byte) error

	// Has the store got key.
	Has(key string) bool

	// Get the key from the store.
	Get(key string) ([]byte, error)

	// Delete the key in the store.
	Del(key string) error
}

type fsStore struct {
	path string
}

func NewFSStore(path string) Store {
	// Get the absolute path to the chosen directory (allows for environment
	// vars and `~`).
	path = paths.AbsPathify(path)

	// Create the filesystem directory.
	err := os.MkdirAll(path, 0750)
	if err != nil {
		log.Fatalf("failed to create fs store: %s", err)
	}

	return &fsStore{path}
}

func (store *fsStore) Set(key string, value []byte) error {
	// Use atomic file writes to prevent partially written files on error.
	return atomicfile.WriteFile(filepath.Join(store.path, key), 0640, bytes.NewReader(value))
}

func (store *fsStore) Has(key string) bool {
	info, err := os.Stat(filepath.Join(store.path, key))
	return !os.IsNotExist(err) && info.Size() > 0
}

func (store *fsStore) Get(key string) ([]byte, error) {
	return os.ReadFile(filepath.Join(store.path, key))
}

func (store *fsStore) Del(key string) error {
	err := os.Remove(filepath.Join(store.path, key))
	return err
}

type memStore struct {
	db map[string][]byte
	mu sync.RWMutex
}

func NewMemStore() Store {
	return &memStore{
		db: make(map[string][]byte),
	}
}

func (store *memStore) Set(key string, value []byte) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.db[key] = value
	return nil
}

func (store *memStore) Has(key string) bool {
	store.mu.RLock()
	defer store.mu.RUnlock()
	val, found := store.db[key]
	return found && len(val) > 0
}

func (store *memStore) Get(key string) ([]byte, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if value, found := store.db[key]; found {
		return value, nil
	}
	return nil, fs.ErrNotExist
}

func (store *memStore) Del(key string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.db, key)
	return nil
}
