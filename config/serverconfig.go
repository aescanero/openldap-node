package config

import (
	"errors"
	"os"
	"strings"
)

type ServerConfig struct {
	AdminPassword       string
	AdminPasswordSHA    string
	AdminPasswordFile   string
	ReplicaPassword     string
	ReplicaPasswordSHA  string
	ReplicaPasswordFile string
	LdapPort            string
	Srvtls              serverTls
	Debug               string
	ApiTls              ApiTlsConfig
	Redis               RedisConfig
	OAuth2              OAuth2Config
}

// ApiTlsConfig holds TLS configuration for the API server
type ApiTlsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
	Port       string `yaml:"port"` // Default 9443 for HTTPS
	AutoRedirect bool `yaml:"auto_redirect"` // Redirect HTTP to HTTPS
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// OAuth2Config holds OAuth2 authentication configuration
type OAuth2Config struct {
	Enabled bool `yaml:"enabled"` // Enable/disable OAuth2 authentication
}

func (scIn *ServerConfig) ImportNotNull(sc *ServerConfig) {
	if sc.AdminPassword != "" {
		scIn.AdminPassword = sc.AdminPassword
	}
	if sc.AdminPasswordFile != "" {
		scIn.AdminPasswordFile = sc.AdminPasswordFile
	}
	if sc.Debug != "" {
		scIn.Debug = sc.Debug
	}
	if sc.LdapPort != "" {
		scIn.LdapPort = sc.LdapPort
	}
	if sc.ReplicaPassword != "" {
		scIn.ReplicaPassword = sc.ReplicaPassword
	}
	if sc.ReplicaPasswordFile != "" {
		scIn.ReplicaPasswordFile = sc.ReplicaPasswordFile
	}
	scIn.Srvtls.ImportNotNull(&sc.Srvtls)
}

func (scIn *ServerConfig) GetAdminPassword() (string, error) {
	if scIn.AdminPassword != "" {
		return scIn.AdminPassword, nil
	}
	if scIn.AdminPasswordFile != "" {
		adminPass, err := os.ReadFile(scIn.AdminPasswordFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSuffix(string(adminPass), "\n"), nil
	}
	return "", errors.New("admin password is required")
}

func (scIn *ServerConfig) GetReplicaPassword() (string, error) {
	if scIn.ReplicaPassword != "" {
		return scIn.ReplicaPassword, nil
	}
	if scIn.ReplicaPasswordFile != "" {
		ReplicaPass, err := os.ReadFile(scIn.ReplicaPasswordFile)
		if err != nil {
			return "", err
		}
		return string(ReplicaPass), nil
	}
	return "", errors.New("replica password is required")
}
