package routes

import (
	"fmt"
	"time"

	"git.sr.ht/~alias/polybase/polybase-http/config"
	"github.com/go-ldap/ldap/v3"
	"github.com/golang-jwt/jwt/v5"
)

func authenticate(username string, password string, cfg *config.Config) (bool, error) {
	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%s", cfg.LDAP.Host, cfg.LDAP.Port))
	if err != nil {
		return false, fmt.Errorf("ldap connect: %w", err)
	}
	defer l.Close()

	l.SetTimeout(5 * time.Second)

	userDN := fmt.Sprintf("cn=%s,%s", username, cfg.LDAP.BaseDN)

	err = l.Bind(userDN, password)
	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return false, nil
		}
		return false, fmt.Errorf("ldap bind: %w", err)
	}

	return true, nil
}

// generateToken creates a JWT token containing just the username
func generateToken(username string, cfg *config.Config) (string, error) {
	type Claims struct {
		Username string `json:"username"`
		jwt.RegisteredClaims
	}

	expiry, err := time.ParseDuration(cfg.Auth.JWTExpiry)
	if err != nil {
		return "", err
	}

	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Auth.JWTSecret))
}
