package cache

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_createMemoryCache(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.CacheConfig
		want        types.Cache
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg:  config.CacheConfig{Type: "memory", Config: map[string]interface{}{}},
		},
		{
			name:        "FailDecodeCfg",
			cfg:         config.CacheConfig{Type: "memory", Config: map[string]interface{}{"default_expiration": []string{}}},
			wantErr:     true,
			errContains: "default_expiration' expected type 'string', got unconvertible type '[]string",
		},
		{
			name:        "FailValidateCfg",
			cfg:         config.CacheConfig{Type: "memory", Config: map[string]interface{}{"default_expiration": ""}},
			wantErr:     true,
			errContains: "Key: 'memoryConfig.DefaultExpiration' Error:Field validation for 'DefaultExpiration' failed on the 'required' tag",
		},
		{
			name:        "FailParseDurationDefaultExpiration",
			cfg:         config.CacheConfig{Type: "memory", Config: map[string]interface{}{"default_expiration": "wrong"}},
			wantErr:     true,
			errContains: "failed to parse default_expiration: time: invalid duration \"wrong\"",
		},
		{
			name:        "FailParseDurationCleanupInterval",
			cfg:         config.CacheConfig{Type: "memory", Config: map[string]interface{}{"cleanup_interval": "wrong"}},
			wantErr:     true,
			errContains: "failed to parse cleanup_interval: time: invalid duration \"wrong\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createMemoryCache(tt.cfg)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}
