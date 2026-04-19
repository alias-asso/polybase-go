package routes

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/alias-asso/polybase-go/polybased/config"
    "github.com/coreos/go-oidc/v3/oidc"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/oauth2"
)

func newOAuth2Config(cfg *config.Config, provider *oidc.Provider) *oauth2.Config {
    return &oauth2.Config{
        ClientID:     cfg.OIDC.ClientID,
        ClientSecret: cfg.OIDC.ClientSecret,
        RedirectURL:  cfg.OIDC.RedirectURI,
        Endpoint:     provider.Endpoint(),
        Scopes:       []string{oidc.ScopeOpenID, "profile"},
    }
}

func generateState() (string, error) {
    b := make([]byte, 16)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

func getOIDCURL(cfg *config.Config, state string) (string, error) {
    ctx := context.Background()
    provider, err := oidc.NewProvider(ctx, cfg.OIDC.IssuerURL)
    if err != nil {
        return "", fmt.Errorf("failed to create OIDC provider: %w", err)
    }

    oauth2Config := newOAuth2Config(cfg, provider)
    return oauth2Config.AuthCodeURL(state), nil
}

func verifyOIDCCode(cfg *config.Config, code string) (string, error) {
    ctx := context.Background()
    provider, err := oidc.NewProvider(ctx, cfg.OIDC.IssuerURL)
    if err != nil {
        return "", fmt.Errorf("failed to create OIDC provider: %w", err)
    }

    oauth2Config := newOAuth2Config(cfg, provider)
    token, err := oauth2Config.Exchange(ctx, code)
    if err != nil {
        return "", fmt.Errorf("failed to exchange code: %w", err)
    }

    rawIDToken, ok := token.Extra("id_token").(string)
    if !ok {
        return "", fmt.Errorf("id_token not found in token response")
    }

    verifier := provider.Verifier(&oidc.Config{ClientID: cfg.OIDC.ClientID})
    idToken, err := verifier.Verify(ctx, rawIDToken)
    if err != nil {
        return "", fmt.Errorf("failed to verify ID token: %w", err)
    }

    var claims struct {
        GivenName string `json:"given_name"`
    }
    if err := idToken.Claims(&claims); err != nil {
        return "", fmt.Errorf("failed to extract claims: %w", err)
    }

    if claims.GivenName == "" {
        return "", fmt.Errorf("given_name claim not found in token")
    }

    return claims.GivenName, nil
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
