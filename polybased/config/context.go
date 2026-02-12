package config

import (
	"context"
)

type key uint8

const (
	cfgKey key = iota
	userKey
)

func CreateContext(ctx context.Context, cfg *Config) context.Context {
	ctx = context.WithValue(ctx, cfgKey, cfg)
	return ctx
}

func SetAuth(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, userKey, username)
}

func GetConfig(ctx context.Context) *Config {
	return ctx.Value(cfgKey).(*Config)
}

func ConnectedUsername(ctx context.Context) string {
	return ctx.Value(userKey).(string)
}
