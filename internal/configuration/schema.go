package configuration

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

type ProviderEnum string

const (
	Discord ProviderEnum = "discord"
)

type Configuration struct {
	ActiveProviders []ProviderEnum

	Discord struct {
		UserID string
		Token  string
	}
}

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
