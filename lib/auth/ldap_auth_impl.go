package auth

import (
	"fmt"
	"github.com/c16a/hermes/lib/config"
	"github.com/go-ldap/ldap/v3"
)

type LdapAuthImpl struct {
	config *config.Config
}

func (impl *LdapAuthImpl) Validate(username string, password string) error {
	authConfig := impl.config.Server.Auth

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", authConfig.LdapHost, authConfig.LdapPort))
	if err != nil {
		return err
	}

	cn := fmt.Sprintf("cn=%s,%s", username, authConfig.LdapDn)
	return l.Bind(cn, password)
}
