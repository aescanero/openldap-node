package ldaputils

import (
	"fmt"
	"log"

	"github.com/aescanero/openldap-node/config"
	"github.com/go-ldap/ldap/v3"
)

func Connect(ldapconfig config.Config, user, pass string) (*ldap.Conn, error) {
	var ldapURL string

	if ldapconfig.SrvConfig.LdapPort != "" {
		ldapURL = "ldap://127.0.0.1:" + ldapconfig.SrvConfig.LdapPort
	} else if ldapconfig.SrvConfig.Srvtls.LdapsPort != "" {
		ldapURL = "ldaps://127.0.0.1:" + ldapconfig.SrvConfig.Srvtls.LdapsPort
	} else {
		log.Fatal("No port defined")
	}

	conn, err := ldap.DialURL(ldapURL)

	if err != nil {
		return nil, fmt.Errorf("Failed to connect. %s", err)
	}

	if err := conn.Bind(user, pass); err != nil {
		return nil, fmt.Errorf("Failed to bind. %s", err)
	}

	return conn, nil
}
