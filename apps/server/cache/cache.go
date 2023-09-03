package cache

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/types"
)

var TypeCache = map[string]CreateCacheFn{}

type CreateCacheFn func(cfg config.CacheConfig) (types.Cache, error)

func CreateCache(cfg config.CacheConfig) (types.Cache, error) {
	if fn, ok := TypeCache[cfg.Type]; ok {
		return fn(cfg)
	}
	return nil, fmt.Errorf("config cache type '%s' does not exist", cfg.Type)
}
