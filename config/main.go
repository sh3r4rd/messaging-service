package config

import (
	"context"
)

type ctxMapKey string

const configKey ctxMapKey = "app_config"

// SaveConfigToContext returns a new context with the provided map stored in it.
func SaveConfigToContext(ctx context.Context, m map[string]string) context.Context {
	return context.WithValue(ctx, configKey, m)
}

// GetConfigFromContext retrieves the map from the context, or nil if not found.
func GetConfigFromContext(ctx context.Context) map[string]string {
	val := ctx.Value(configKey)
	if m, ok := val.(map[string]string); ok {
		return m
	}
	return nil
}

func GetValueFromConfig(ctx context.Context, key string) (string, bool) {
	m := GetConfigFromContext(ctx)
	if m == nil {
		return "", false
	}
	value, exists := m[key]
	return value, exists
}
