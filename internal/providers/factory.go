package providers

import (
	"errors"

	"github.com/erykksc/notifr/internal/configuration"
)

func CreateProvider(config configuration.Configuration) (MsgProvider, error) {
	var providers []MsgProvider

	for _, provider := range config.ActiveProviders {
		switch provider {
		case configuration.Discord:
			dConf := config.Discord
			discordP := CreateDiscord(dConf.Token, dConf.UserID)
			providers = append(providers, &discordP)
		default:
			return nil, errors.New("unsupported provider: " + string(provider))
		}
	}

	if len(providers) == 0 {
		return nil, errors.New("no active providers specified")
	}

	return providers[0], nil
}
