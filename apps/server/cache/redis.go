package cache

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/mapstructure"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/eko/gocache/lib/v4/cache"
	gocacheStore "github.com/eko/gocache/store/redis/v4"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

const (
	redisKey = "redis"
)

func init() {
	TypeCache[redisKey] = createRedisCache
}

type redisConfig struct {
	Address  string `mapstructure:"address" validate:"required"`
	DB       int    `mapstructure:"db"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func createRedisCache(cfg config.CacheConfig) (types.Cache, error) {
	instanceConfig := redisConfig{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	clientConfig := &redis.Options{
		Addr: instanceConfig.Address,
	}

	if instanceConfig.DB != 0 {
		clientConfig.DB = instanceConfig.DB
	}

	if instanceConfig.Username != "" {
		clientConfig.Username = instanceConfig.Username
	}

	if instanceConfig.Password != "" {
		clientConfig.Password = instanceConfig.Password
	}

	cacheClient := redis.NewClient(clientConfig)
	cacheStore := gocacheStore.NewRedis(cacheClient)
	cacheManager := cache.New[string](cacheStore)
	return cacheManager, nil
}
