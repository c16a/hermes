package auth

import (
	"errors"
	"github.com/c16a/hermes/lib/config"
)

type AuthorisationProvider interface {
	Validate(string, string) error
}

func FetchProviderFromConfig(config *config.Config) (provider AuthorisationProvider, err error) {
	authConfig := config.Server.Auth

	if authConfig == nil || len(authConfig.Type) == 0 {
		return
	}

	switch authConfig.Type {
	case "ldap":
		provider = &LdapAuthImpl{config: config}
		break
	default:
		err = errors.New("no valid auth provider found")
	}

	return
}
