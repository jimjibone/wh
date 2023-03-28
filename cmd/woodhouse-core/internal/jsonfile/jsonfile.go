package jsonfile

import (
	"encoding/json"
	"io"
	"os"
)

func LoadFile(data interface{}, filename string) error {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		// Open the file.
		f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		// Decode the config.
		err = json.NewDecoder(f).Decode(data)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

func SaveFile(data interface{}, filename string) error {
	// Open/create the file.
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Encode the config.
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func SaveFileIfNotExist(data interface{}, filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return SaveFile(data, filename)
	}
	return nil
}
