package yamlfile

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/jimjibone/woodhouse-4/shared/atomicfile"
	"gopkg.in/yaml.v3"
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
	// Encode the config.
	buf := &bytes.Buffer{}
	err := yaml.NewEncoder(buf).Encode(data)
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
