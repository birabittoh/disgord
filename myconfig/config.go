package myconfig

import (
	"encoding/json"
	"os"
)

type Config[T any] struct {
	filename string
	Values   T
}

func New[T any](filename string) (c *Config[T], err error) {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	c = &Config[T]{
		filename: filename,
	}

	err = json.Unmarshal(fileContent, &c.Values)
	return
}

func (c *Config[T]) Save() error {
	fileContent, err := json.MarshalIndent(c.Values, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.filename, fileContent, 0644)
}
