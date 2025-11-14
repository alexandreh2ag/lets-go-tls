package cache

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/mapstructure"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/eko/gocache/lib/v4/cache"
	gocacheStore "github.com/eko/gocache/store/go_cache/v4"
	"github.com/go-playground/validator/v10"
	go_cache "github.com/patrickmn/go-cache"
	"time"
)

const (
	memoryKey = "memory"
)

func init() {
	TypeCache[memoryKey] = createMemoryCache
}

type memoryConfig struct {
	DefaultExpiration string `mapstructure:"default_expiration" validate:"required"`
	CleanupInterval   string `mapstructure:"cleanup_interval" validate:"required"`
}

func createMemoryCache(cfg config.CacheConfig) (types.Cache, error) {
	instanceConfig := memoryConfig{
		DefaultExpiration: "5m",
		CleanupInterval:   "10m",
	}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}
	defaultExpiration, err := time.ParseDuration(instanceConfig.DefaultExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse default_expiration: %v", err)
	}
	cleanupInterval, err := time.ParseDuration(instanceConfig.CleanupInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cleanup_interval: %v", err)
	}
	cacheClient := go_cache.New(defaultExpiration, cleanupInterval)
	cacheStore := gocacheStore.NewGoCache(cacheClient)
	cacheManager := cache.New[string](cacheStore)
	return cacheManager, nil
}
