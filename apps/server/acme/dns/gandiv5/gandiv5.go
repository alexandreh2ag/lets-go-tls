package gandiv5

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/mapstructure"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/providers/dns/gandiv5"
	"github.com/go-playground/validator/v10"
	"time"
)

const KeyDnsGandiV5 = "gandiv5"

type ConfigGandiV5 struct {
	APIKey string `mapstructure:"api_key" validate:"required"`

	PropagationTimeout time.Duration `mapstructure:"propagation_timeout" validate:"required"`
	PollingInterval    time.Duration `mapstructure:"polling_interval" validate:"required"`
	HttpTimeout        time.Duration `mapstructure:"http_timeout" validate:"required"`
}

type gandiV5 struct {
	*gandiv5.DNSProvider
	id string
}

func (g *gandiV5) ID() string {
	return fmt.Sprintf("%s-%s", KeyDnsGandiV5, g.id)
}

func (g *gandiV5) Type() string {
	return acme.TypeDNS01
}

func CreateGandiV5(_ *context.ServerContext, id string, cfg map[string]interface{}) (acme.Challenge, error) {
	config := gandiv5.NewDefaultConfig()
	instanceConfig := ConfigGandiV5{
		PropagationTimeout: config.PropagationTimeout,
		PollingInterval:    config.PollingInterval,
		HttpTimeout:        config.HTTPClient.Timeout,
	}
	err := mapstructure.Decode(cfg, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	config.PersonalAccessToken = instanceConfig.APIKey
	provider, err := gandiv5.NewDNSProviderConfig(config)
	return &gandiV5{id: id, DNSProvider: provider}, err
}
