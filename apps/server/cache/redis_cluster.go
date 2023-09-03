package cache

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/eko/gocache/lib/v4/cache"
	gocacheStore "github.com/eko/gocache/store/redis/v4"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

const (
	redisClusterKey = "redis-cluster"
)

func init() {
	TypeCache[redisClusterKey] = createRedisClusterCache
}

type redisClusterConfig struct {
	Address  []string `mapstructure:"address" validate:"required,min=1"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
}

func createRedisClusterCache(cfg config.CacheConfig) (types.Cache, error) {
	instanceConfig := redisClusterConfig{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	clientConfig := &redis.ClusterOptions{
		Addrs: instanceConfig.Address,
	}

	if instanceConfig.Username != "" {
		clientConfig.Username = instanceConfig.Username
	}

	if instanceConfig.Password != "" {
		clientConfig.Password = instanceConfig.Password
	}

	cacheClient := redis.NewClusterClient(clientConfig)
	cacheStore := gocacheStore.NewRedis(cacheClient)
	cacheManager := cache.New[string](cacheStore)
	return cacheManager, nil
}
