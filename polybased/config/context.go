package config

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type key uint8

const (
	cfgKey   key = 0
	userKey  key = 1
	loggedIn key = 2
	devMode  key = 3
)

func CreateContext(ctx context.Context, cfg *Config, dev bool) context.Context {
	ctx = context.WithValue(ctx, cfgKey, cfg)
	ctx = context.WithValue(ctx, devMode, dev)
	return ctx
}

func SetAuth(ctx context.Context, r *http.Request) context.Context {
	logged, username := userConnected(GetConfig(ctx), r)
	ctx = context.WithValue(ctx, loggedIn, logged)
	if logged {
		ctx = context.WithValue(ctx, userKey, username)
	}
	return ctx
}

func userConnected(cfg *Config, r *http.Request) (bool, string) {
	cookie, err := r.Cookie("X-Auth-Token")
	if err != nil {
		return false, ""
	}

	type Claims struct {
		Username string `json:"username"`
		jwt.RegisteredClaims
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(cfg.Auth.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return false, ""
	}

	v, ok := token.Claims.(*Claims)
	return ok, v.Username
}

func GetConfig(ctx context.Context) *Config {
	return ctx.Value(cfgKey).(*Config)
}

func IsLogged(ctx context.Context) bool {
	return ctx.Value(loggedIn).(bool)
}

func GetUsername(ctx context.Context) string {
	return ctx.Value(userKey).(string)
}

func IsDev(ctx context.Context) bool {
	return ctx.Value(devMode).(bool)
}
