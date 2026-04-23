package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
)

type Server struct {
	Host string
	Port string
	Log  string
}

type Database struct {
	Path string
}

type OIDC struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	IssuerURL    string `toml:"issuer_url"`
	RedirectURI  string `toml:"redirect_uri"`
}

type Auth struct {
	JWTSecret string `toml:"jwt_secret"`
	JWTExpiry string `toml:"jwt_expiry"`
}

type Config struct {
	Server   Server
	Database Database
	OIDC     OIDC
	Auth     Auth
}

func DefaultConfig() Config {
	return Config{
		Server: Server{
			Host: "127.0.0.1",
			Port: "1265",
			Log:  "/var/log/polybase/polybase.log",
		},
		Database: Database{
			Path: "/var/lib/polybase/polybase.db",
		},
		Auth: Auth{
			JWTExpiry: "72h",
		},
	}
}

func LoadConfig(configPath string) (Config, error) {
	config := DefaultConfig()
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return Config{}, err
	}

	config.loadFromEnv()

	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func (c *Config) loadFromEnv() {
	if host := os.Getenv("POLYBASE_SERVER_HOST"); host != "" {
		c.Server.Host = host
	}
	if port := os.Getenv("POLYBASE_SERVER_PORT"); port != "" {
		c.Server.Port = port
	}
	if log := os.Getenv("POLYBASE_SERVER_LOG"); log != "" {
		c.Server.Log = log
	}

	if path := os.Getenv("POLYBASE_DATABASE_PATH"); path != "" {
		c.Database.Path = path
	}

	if clientID := os.Getenv("POLYBASE_OIDC_CLIENT_ID"); clientID != "" {
		c.OIDC.ClientID = clientID
	}
	if clientSecret := os.Getenv("POLYBASE_OIDC_CLIENT_SECRET"); clientSecret != "" {
		c.OIDC.ClientSecret = clientSecret
	}
	if issuerURL := os.Getenv("POLYBASE_OIDC_ISSUER_URL"); issuerURL != "" {
		c.OIDC.IssuerURL = issuerURL
	}
	if redirectURI := os.Getenv("POLYBASE_OIDC_REDIRECT_URI"); redirectURI != "" {
		c.OIDC.RedirectURI = redirectURI
	}

	if secret := os.Getenv("POLYBASE_AUTH_JWT_SECRET"); secret != "" {
		c.Auth.JWTSecret = secret
	}
	if expiry := os.Getenv("POLYBASE_AUTH_JWT_EXPIRY"); expiry != "" {
		c.Auth.JWTExpiry = expiry
	}
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

	if c.Server.Log == "" {
		return fmt.Errorf("server.log is required")
	}

	// Database validation
	if c.Database.Path == "" {
		return fmt.Errorf("database.path is required")
	}

	// OIDC validation
	if c.OIDC.ClientID == "" {
		return fmt.Errorf("oidc.client_id is required")
	}
	if c.OIDC.ClientSecret == "" {
		return fmt.Errorf("oidc.client_secret is required")
	}
	if c.OIDC.IssuerURL == "" {
		return fmt.Errorf("oidc.issuer_url is required")
	}
	if c.OIDC.RedirectURI == "" {
		return fmt.Errorf("oidc.redirect_uri is required")
	}

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
