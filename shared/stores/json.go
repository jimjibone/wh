package stores

import "encoding/json"

func SetJson(store Store, key string, data any) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return store.Set(key, raw)
}

func GetJson(store Store, key string, data any) error {
	raw, err := store.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, data)
}
