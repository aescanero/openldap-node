package ldaputils

import (
	"log"

	"github.com/go-ldap/ldap/v3"
)

func GetOne(conn *ldap.Conn, baseDN string, atributes ...string) (map[string]string, error) {

	filterDN := "(objectclass=*)"

	result, err := conn.Search(ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filterDN,
		atributes,
		nil,
	))
	if err != nil {
		return nil, err
	}

	values := make(map[string]string)

	for _, entry := range result.Entries {
		for _, attr := range entry.Attributes {
			values[attr.Name] = attr.Values[0]
			log.Printf("Name: %s,Values: %s", attr.Name, values[attr.Name])
		}
	}

	return values, nil
}
