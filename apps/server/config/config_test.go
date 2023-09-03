package config

import (
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/go-acme/lego/v4/lego"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	got := NewConfig()
	assert.Equal(t, Config{}, got)
}

func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig()
	assert.Equal(t,
		Config{
			HTTP:                    config.HTTPConfig{Listen: "0.0.0.0:8080"},
			Interval:                time.Minute * 5,
			LockDuration:            time.Minute * 25,
			UnusedRetentionDuration: time.Hour * 24 * 14,
			Cache:                   CacheConfig{Type: "memory"},
			Acme: AcmeConfig{
				CAServer:    lego.LEDirectoryProduction,
				RenewPeriod: time.Hour * 24 * 10,
				Resolvers:   map[string]ResolverConfig{},
				MaxAttempt:  3,
				DelayFailed: time.Hour * 24,
			},
			JWT: JWTConfig{Method: "HS256"},
		},
		got,
	)
}
