package stores

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFSStoreSetGetHasDel(t *testing.T) {
	dir := t.TempDir()
	store := NewFSStore(dir)

	const key = "wh.token"
	value := []byte("super-secret-refresh-token")

	if store.Has(key) {
		t.Fatalf("expected store to not have %q before Set", key)
	}

	if err := store.Set(key, value); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if !store.Has(key) {
		t.Fatalf("expected store to have %q after Set", key)
	}

	got, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("Get() = %q, want %q", got, value)
	}

	if err := store.Del(key); err != nil {
		t.Fatalf("Del() error = %v", err)
	}
	if store.Has(key) {
		t.Fatalf("expected store to not have %q after Del", key)
	}
}

func TestNewFSStoreTightensPreExistingModes(t *testing.T) {
	dir := t.TempDir()

	// Pre-seed a looser directory containing a looser file, as if it had
	// been created by an older version of this library (or by hand).
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatalf("failed to seed dir mode: %v", err)
	}

	filePath := filepath.Join(dir, "wh.cert")
	if err := os.WriteFile(filePath, []byte("legacy pinned cert"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	NewFSStore(dir)

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("failed to stat dir: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0700 {
		t.Fatalf("dir mode = %o, want %o", perm, 0700)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0600 {
		t.Fatalf("file mode = %o, want %o", perm, 0600)
	}
}

func TestFSStoreSetTightensPreExistingFileMode(t *testing.T) {
	dir := t.TempDir()
	store := NewFSStore(dir)

	const key = "wh.cert"
	filePath := filepath.Join(dir, key)

	// Simulate a file that was left behind with a looser mode.
	if err := os.WriteFile(filePath, []byte("old value"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	if err := store.Set(key, []byte("new value")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("file mode after Set() = %o, want %o", perm, 0600)
	}
}
