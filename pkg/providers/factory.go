package providers

import (
	"errors"

	"github.com/erykksc/chatreply/pkg/configuration"
)

// Every MsgProvider should have a function that implement this signature
type ProviderFactoryFunc func(configuration.Configuration) (MsgProvider, error)

// CreateProvider is a MsgProvider factory function
func CreateProvider(config configuration.Configuration) (MsgProvider, error) {
	activeProvider := config.ActiveProvider
	if activeProvider == "" {
		return nil, errors.New("no active providers specified")
	}

	providers := make(map[string]ProviderFactoryFunc)

	// Register all providers
	providers["discord"] = CreateDiscord
	providers["telegram"] = CreateTelegram

	createProviderFunc, ok := providers[activeProvider]
	if !ok {
		return nil, errors.New("unsupported provider: " + activeProvider)
	}

	provider, err := createProviderFunc(config)
	if err != nil {
		return nil, errors.New("error creating provider: " + err.Error())
	}

	return provider, nil
}
