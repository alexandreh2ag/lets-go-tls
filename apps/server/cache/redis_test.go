package cache

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_createRedisCache(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.CacheConfig
		want        types.Cache
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg:  config.CacheConfig{Type: "redis", Config: map[string]interface{}{"address": "127.0.0.1:6379"}},
		},
		{
			name: "SuccessWithConfig",
			cfg:  config.CacheConfig{Type: "redis", Config: map[string]interface{}{"address": "127.0.0.1:6379", "db": 2, "username": "user", "password": "pass"}},
		},
		{
			name:        "FailDecodeCfg",
			cfg:         config.CacheConfig{Type: "redis", Config: map[string]interface{}{"address": []string{}}},
			wantErr:     true,
			errContains: "'address' expected type 'string', got unconvertible type '[]string",
		},
		{
			name:        "FailValidateCfg",
			cfg:         config.CacheConfig{Type: "redis", Config: map[string]interface{}{"address": ""}},
			wantErr:     true,
			errContains: "Key: 'redisConfig.Address' Error:Field validation for 'Address' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createRedisCache(tt.cfg)

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
