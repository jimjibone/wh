package wh

import (
	"fmt"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/shared/stores"
)

type clientStore struct {
	store stores.Store
}

func newClientStore(store stores.Store) *clientStore {
	return &clientStore{
		store: store,
	}
}

// Upgrade the store to the latest schema.
func (store *clientStore) Upgrade(log *log.Context) error {
	renamer := func(from, to string) error {
		if store.store.Has(from) {
			log.Debugf("store renaming %q to %q", from, to)
			value, err := store.store.Get(from)
			if err != nil {
				return fmt.Errorf("get %q: %s", from, err)
			}
			store.store.Del(from)

			err = store.store.Set(to, value)
			if err != nil {
				return fmt.Errorf("set %q: %s", to, err)
			}
		}
		return nil
	}

	if err := renamer("id", "wh.id"); err != nil {
		return err
	}
	if err := renamer("cert", "wh.crt"); err != nil {
		return err
	}
	if err := renamer("token", "wh.token"); err != nil {
		return err
	}
	return nil
}

func (store *clientStore) HasID() bool            { return store.store.Has("wh.id") }
func (store *clientStore) GetID() ([]byte, error) { return store.store.Get("wh.id") }
func (store *clientStore) SetID(v []byte) error   { return store.store.Set("wh.id", v) }

func (store *clientStore) HasCert() bool            { return store.store.Has("wh.cert") }
func (store *clientStore) GetCert() ([]byte, error) { return store.store.Get("wh.cert") }
func (store *clientStore) SetCert(v []byte) error   { return store.store.Set("wh.cert", v) }
func (store *clientStore) DelCert() error           { return store.store.Del("wh.cert") }

func (store *clientStore) HasToken() bool            { return store.store.Has("wh.token") }
func (store *clientStore) GetToken() ([]byte, error) { return store.store.Get("wh.token") }
func (store *clientStore) SetToken(v []byte) error   { return store.store.Set("wh.token", v) }
func (store *clientStore) DelToken() error           { return store.store.Del("wh.token") }
