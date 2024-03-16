package jsonfile

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/jimjibone/woodhouse-4/shared/atomicfile"
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
	// Encode the config.
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	err := enc.Encode(data)
	if err != nil {
		return err
	}

	// Atomically write the file.
	return atomicfile.WriteFile(filename, 0644, buf)
}

func SaveFileIfNotExist(data interface{}, filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return SaveFile(data, filename)
	}
	return nil
}
