package clients

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/shared/stores"
)

// TestClientStoreUpgradeMigratesCertKey exercises the migration path the way
// production reaches it: a bridge store created before the "wh."-prefixed
// key scheme existed has a legacy "cert" file on disk, and Upgrade (invoked
// from Client.Run) must rename it to the key the accessors actually use
// ("wh.cert"), not "wh.crt".
func TestClientStoreUpgradeMigratesCertKey(t *testing.T) {
	dir := t.TempDir()

	// Seed a legacy pinned cert directly on disk, bypassing the store's own
	// Set(), since that's how a real pre-migration store would look.
	legacyValue := []byte("legacy pinned server cert")
	if err := os.WriteFile(filepath.Join(dir, "cert"), legacyValue, 0644); err != nil {
		t.Fatalf("failed to seed legacy cert file: %v", err)
	}

	fsStore := stores.NewFSStore(dir)
	store := newClientStore(fsStore)

	logCtx := log.NewContext(log.DefaultLogger, "test", log.WarnLevel)
	if err := store.Upgrade(logCtx); err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}

	if !store.HasCert() {
		t.Fatalf("expected migrated cert to be retrievable via wh.cert accessor")
	}

	got, err := store.GetCert()
	if err != nil {
		t.Fatalf("GetCert() error = %v", err)
	}
	if string(got) != string(legacyValue) {
		t.Fatalf("GetCert() = %q, want %q", got, legacyValue)
	}

	if _, err := os.Stat(filepath.Join(dir, "cert")); !os.IsNotExist(err) {
		t.Fatalf("expected legacy %q file to be gone, stat err = %v", "cert", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "wh.crt")); !os.IsNotExist(err) {
		t.Fatalf("migration must not create the old buggy %q key", "wh.crt")
	}
}
