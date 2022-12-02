package yamlfile

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
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
		err = yaml.NewDecoder(f).Decode(data)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if te, ok := err.(*yaml.TypeError); ok {
				fmt.Println(te.Errors)
			}
			// fmt.Println(yaml.FormatError(err, true, true))
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
	err = yaml.NewEncoder(f).Encode(data)
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
