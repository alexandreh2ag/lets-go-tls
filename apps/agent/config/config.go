package config

import (
	"github.com/alexandreh2ag/lets-go-tls/config"
	"time"
)

type Config struct {
	Requesters []config.RequesterConfig `mapstructure:"requesters" validate:"required,unique=Id,dive"`
	Storages   []StorageConfig          `mapstructure:"storages" validate:"required,unique=Id,dive"`
	State      config.StateConfig       `mapstructure:"state" validate:"required"`
	HTTP       config.HTTPConfig        `mapstructure:"http" validate:"required"`
	Interval   time.Duration            `mapstructure:"interval" validate:"required"`
	Manager    ManagerConfig            `mapstructure:"manager" validate:"required"`
}
type ManagerConfig struct {
	Address  string `mapstructure:"address" validate:"required,http_url"`
	TokenJWT string `mapstructure:"token" validate:"required"`
}

func NewConfig() Config {
	return Config{}
}

func DefaultConfig() Config {
	cfg := NewConfig()
	cfg.Interval = time.Minute * 5
	cfg.HTTP = config.HTTPConfig{Listen: "0.0.0.0:8080"}
	return cfg
}
