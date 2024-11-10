package config

import (
	"github.com/BurntSushi/toml"
)

// Config represents server configuration
type Config struct {
	Server struct {
		Host string
		Port string
	}
	Database struct {
		Path string
	}
	LDAP struct {
		Host       string
		Port       string
		BaseDN     string
		UserDN     string
		AdminGroup string
	}
	Auth struct {
		JWTSecret string `toml:"jwt_secret"`
		JWTExpiry string `toml:"jwt_expiry"`
	}
}

func LoadConfig(configPath string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return Config{}, err
	}

	// TODO: validate the config and fill it with proper defaults

	return config, nil
}
