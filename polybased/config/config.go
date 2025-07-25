package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-ldap/ldap/v3"
)

type Server struct {
	Host string
	Port string
	Mode string
	Log  string
}

type Database struct {
	Path string
}

type LDAP struct {
	Host   string `toml:"host"`
	Port   string `toml:"port"`
	UserDN string `toml:"user_dn"`
}

type Auth struct {
	JWTSecret string `toml:"jwt_secret"`
	JWTExpiry string `toml:"jwt_expiry"`
}

type Config struct {
	Server   Server
	Database Database
	LDAP     LDAP
	Auth     Auth
}

func DefaultConfig() Config {
	return Config{
		Server: Server{
			Host: "127.0.0.1",
			Port: "1265",
			Mode: "prod",
			Log:  "/var/log/polybase/polybase.log",
		},
		Database: Database{
			Path: "/var/lib/polybase/polybase.db",
		},
		Auth: Auth{
			JWTExpiry: "4320h",
		},
	}
}

func LoadConfig(configPath string) (Config, error) {
	config := DefaultConfig()
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return Config{}, err
	}

	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	// Server validation
	if c.Server.Host == "" {
		return fmt.Errorf("server.host is required")
	}
	if ip := net.ParseIP(c.Server.Host); ip == nil && c.Server.Host != "0.0.0.0" {
		return fmt.Errorf("server.host must be a valid IP address")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}
	if port, err := strconv.Atoi(c.Server.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("server.port must be a valid port number (1-65535)")
	}

	if c.Server.Mode == "" {
		return fmt.Errorf("server.mode is required")
	}

	if c.Server.Log == "" {
		return fmt.Errorf("server.log is required")
	}

	// Database validation
	if c.Database.Path == "" {
		return fmt.Errorf("database.path is required")
	}

	// LDAP validation
	if c.LDAP.Host == "" {
		return fmt.Errorf("ldap.host is required")
	}
	if c.LDAP.Port == "" {
		return fmt.Errorf("ldap.port is required")
	}
	if port, err := strconv.Atoi(c.LDAP.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("ldap.port must be a valid port number (1-65535)")
	}
	if c.LDAP.UserDN == "" {
		return fmt.Errorf("ldap.user_dn is required")
	}
	if !strings.Contains(c.LDAP.UserDN, "%s") {
		return fmt.Errorf("ldap.user_dn must contain %%s placeholder for username")
	}

	// Test LDAP connection
	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%s", c.LDAP.Host, c.LDAP.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer l.Close()

	// Auth validation
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	if len(c.Auth.JWTSecret) < 16 {
		return fmt.Errorf("auth.jwt_secret should be at least 16 characters long")
	}

	if c.Auth.JWTExpiry == "" {
		return fmt.Errorf("auth.jwt_expiry is required")
	}
	if _, err := time.ParseDuration(c.Auth.JWTExpiry); err != nil {
		return fmt.Errorf("auth.jwt_expiry must be a valid duration (e.g., '24h', '168h')")
	}

	return nil
}
