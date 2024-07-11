package configuration

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	ActiveProvider string

	Discord struct {
		UserID string
		Token  string
	}
}

// LoadConfiguration loads a configuration from a .toml file
func LoadConfiguration(path string) (Configuration, error) {
	config := Configuration{}

	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()

	fData, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}

	err = toml.Unmarshal(fData, &config)

	return config, err
}
