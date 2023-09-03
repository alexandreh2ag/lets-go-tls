package config

import (
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/go-acme/lego/v4/lego"
	"time"
)

type Config struct {
	Requesters              []config.RequesterConfig `mapstructure:"requesters" validate:"required,unique=Id,dive"`
	Acme                    AcmeConfig               `mapstructure:"acme" validate:"required"`
	State                   config.StateConfig       `mapstructure:"state" validate:"required"`
	Cache                   CacheConfig              `mapstructure:"cache" validate:"required"`
	HTTP                    config.HTTPConfig        `mapstructure:"http" validate:"required"`
	JWT                     JWTConfig                `mapstructure:"jwt" validate:"required"`
	Interval                time.Duration            `mapstructure:"interval" validate:"required"`
	LockDuration            time.Duration            `mapstructure:"lock_duration" validate:"required"`
	UnusedRetentionDuration time.Duration            `mapstructure:"unused_retention" validate:"required"`
}

type AcmeConfig struct {
	CAServer    string                    `mapstructure:"ca_server" validate:"required"`
	Email       string                    `mapstructure:"email" validate:"required,email"`
	Resolvers   map[string]ResolverConfig `mapstructure:"resolvers,omitempty"`
	RenewPeriod time.Duration             `mapstructure:"renew_period" validate:"required"`
	MaxAttempt  int                       `mapstructure:"max_attempt" validate:"required,min=1"`
	DelayFailed time.Duration             `mapstructure:"delay_failed" validate:"required"`
}

type CacheConfig struct {
	Type   string                 `mapstructure:"type" validate:"required,excludesall=!@#$ "`
	Config map[string]interface{} `mapstructure:"config,omitempty"`
}

type ResolverConfig struct {
	Type    string                 `mapstructure:"type" validate:"required,excludesall=!@#$ "`
	Config  map[string]interface{} `mapstructure:"config"`
	Filters []string               `mapstructure:"filters" validate:"required,min=1"`
}

type JWTConfig struct {
	Key    string `mapstructure:"key" validate:"required,min=1"`
	Method string `mapstructure:"method" validate:"required"`
}

func NewConfig() Config {
	return Config{}
}

func DefaultConfig() Config {
	cfg := NewConfig()
	cfg.Interval = time.Minute * 5
	cfg.LockDuration = time.Minute * 25
	cfg.UnusedRetentionDuration = time.Hour * 24 * 14
	cfg.HTTP = config.HTTPConfig{Listen: "0.0.0.0:8080"}
	cfg.Cache = CacheConfig{Type: "memory"}
	cfg.Acme = AcmeConfig{
		CAServer:    lego.LEDirectoryProduction,
		Resolvers:   map[string]ResolverConfig{},
		RenewPeriod: time.Hour * 24 * 10,
		MaxAttempt:  3,
		DelayFailed: time.Hour * 24,
	}
	cfg.JWT = JWTConfig{Method: "HS256"}
	return cfg
}
