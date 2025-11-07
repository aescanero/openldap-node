package ldaputils

import (
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/aescanero/openldap-node/config"
	"github.com/aescanero/openldap-node/utils"
	"github.com/go-ldap/ldap/v3"
)

//go:embed templates/slapd.conf.tmpl
var slapdConfTemplate string

//go:embed templates/base.ldif.tmpl
var baseLdifTemplate string

func Prepare(myConfig config.Config) error {

	//var conf embed.FS
	base := myConfig.Database[0].Base
	adminPassword, err := myConfig.SrvConfig.GetAdminPassword()
	if err != nil {
		log.Fatal("Error loading config:" + err.Error())
		panic(err)
	}

	encode := utils.Encode{}
	adminPasswordSHA := encode.MakeSSHAEncode([]byte(adminPassword))
	myConfig.SrvConfig.AdminPasswordSHA = fmt.Sprintf("{SSHA}%s", base64.StdEncoding.EncodeToString(adminPasswordSHA))
	fmt.Printf("passwordSSHAB64: %s\n", myConfig.SrvConfig.AdminPasswordSHA)
	if myConfig.Database[0].Replicatls[0].ReplicaUrl == "" {
		myConfig.Database[0].Replicatls = nil
	}

	slapdConf, err := template.New("slapdConf").Parse(slapdConfTemplate) //template.ParseFS(conf, "templates/slapd.conf.tmpl")
	if err != nil {
		log.Fatal("Error loading templates:" + err.Error())
		panic(err)
	}

	err = utils.CreateDirs([]string{"/etc/ldap", "/etc/ldap/slapd.d", "/var/lib/ldap/0", "/etc/ldap/schema", "/etc/ldap/certs"})
	if err != nil {
		log.Println(err)
		panic(err)
	}

	schemaFiles := []string{"/etc/openldap/schema/core.schema",
		"/etc/openldap/schema/cosine.schema",
		"/etc/openldap/schema/misc.schema",
		"/etc/openldap/schema/inetorgperson.schema",
		"/etc/openldap/schema/nis.schema"}

	for _, schema := range myConfig.Schemas {
		schemaFiles = append(schemaFiles, schema.Path)
	}

	err = utils.CopyFiles(
		schemaFiles,
		"/etc/ldap/schema",
	)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	if myConfig.SrvConfig.Srvtls.LdapsTls.CaFile != "" && myConfig.SrvConfig.Srvtls.LdapsTls.CrtFile != "" && myConfig.SrvConfig.Srvtls.LdapsTls.CrtKeyFile != "" {
		tlsfiles := []string{myConfig.SrvConfig.Srvtls.LdapsTls.CaFile, myConfig.SrvConfig.Srvtls.LdapsTls.CrtFile, myConfig.SrvConfig.Srvtls.LdapsTls.CrtKeyFile}
		err = utils.CopyFiles(
			tlsfiles,
			"/etc/ldap/certs",
		)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		caFilename := filepath.Base(myConfig.SrvConfig.Srvtls.LdapsTls.CaFile)
		crtFilename := filepath.Base(myConfig.SrvConfig.Srvtls.LdapsTls.CrtFile)
		crtKeyFilename := filepath.Base(myConfig.SrvConfig.Srvtls.LdapsTls.CrtKeyFile)
		myConfig.SrvConfig.Srvtls.LdapsTls.CaFile = "/etc/ldap/certs/" + caFilename
		myConfig.SrvConfig.Srvtls.LdapsTls.CrtFile = "/etc/ldap/certs/" + crtFilename
		myConfig.SrvConfig.Srvtls.LdapsTls.CrtKeyFile = "/etc/ldap/certs/" + crtKeyFilename
	} else if myConfig.SrvConfig.Srvtls.LdapsTls.CaFile != "" || myConfig.SrvConfig.Srvtls.LdapsTls.CrtFile != "" || myConfig.SrvConfig.Srvtls.LdapsTls.CrtKeyFile != "" {
		panic(errors.New("the cafile, crtfile, crtkeyfile must be set to obtain TLS support"))
	}

	f, err := os.Create("/tmp/slapd.conf")
	if err != nil {
		log.Print("Can't create ", "/tmp/slapd.conf")
		panic(err)
	}

	err = slapdConf.Execute(io.Writer(f), myConfig)
	if err != nil {
		log.Print("Can't execute ", "templates/slapd.conf.tmpl")
		panic(err)
	}

	log.Print("Populating slapd conf")
	///usr/sbin/slaptest -f /tmp/slapd.conf -F /etc/ldap/slapd.d
	out, _ := exec.Command("/usr/sbin/slaptest", "-f", "/tmp/slapd.conf", "-F", "/etc/ldap/slapd.d").Output()
	log.Printf("RES %s\n", out)

	log.Print("Initializing database")
	f, err = os.Create("/tmp/base.ldif")
	if err != nil {
		log.Print("Can't create ", "/tmp/base.ldif")
		panic(err)
	}

	baseLdifTemplate = `dn: ou=templates,` + base + `
objectClass: organizationalUnit
ou: templates` + "\n\n" + baseLdifTemplate

	parsedDN, err := ldap.ParseDN(base)
	if err != nil || len(parsedDN.RDNs) == 0 {
		log.Println(err)
		panic(err)
	}
	switch parsedDN.RDNs[0].Attributes[0].Type {
	case "o":
		baseLdifTemplate = `dn: ` + base + `
objectClass: organization
o: ` + parsedDN.RDNs[0].Attributes[0].Value + "\n\n" + baseLdifTemplate

	case "dc":
		baseLdifTemplate = `dn: ` + base + `
objectClass: dcObject
objectClass: organization
dc: ` + parsedDN.RDNs[0].Attributes[0].Value + `
o: ` + parsedDN.RDNs[0].Attributes[0].Value + "\n\n" + baseLdifTemplate
	}

	baseLdap, err := template.New("baseLdap").Parse(baseLdifTemplate) //template.ParseFS(conf, "templates/base.ldif.tmpl")
	if err != nil {
		log.Fatal("Error loading templates:" + err.Error())
		panic(err)
	}

	config := map[string]string{
		"ldapRoot": base,
	}

	err = baseLdap.Execute(io.Writer(f), config)
	if err != nil {
		log.Print("Can't execute ", "/tmp/base.ldif")
		panic(err)
	}

	_, err = exec.Command("/usr/sbin/slapadd", "-F", "/etc/ldap/slapd.d", "-l", "/tmp/base.ldif").Output()
	if err != nil {
		log.Fatal("Can't execute /usr/sbin/slapadd -F /etc/ldap/slapd.d -l /tmp/base.ldif")
		panic(err)
	}

	log.Print("Configuring Openldap")
	return nil
}
