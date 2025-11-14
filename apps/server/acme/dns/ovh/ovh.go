package ovh

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/mapstructure"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	legoOVH "github.com/go-acme/lego/v4/providers/dns/ovh"
	"time"
)

const KeyDnsOVH = "ovh"

type ConfigOvh struct {
	AccessToken       string `mapstructure:"access_token"`
	ApplicationKey    string `mapstructure:"application_key"`
	ApplicationSecret string `mapstructure:"application_secret"`
	ClientID          string `mapstructure:"client_id"`
	ClientSecret      string `mapstructure:"client_secret"`
	ConsumerKey       string `mapstructure:"consumer_key"`
	Endpoint          string `mapstructure:"endpoint"`

	PropagationTimeout time.Duration `mapstructure:"propagation_timeout" validate:"required"`
	PollingInterval    time.Duration `mapstructure:"polling_interval" validate:"required"`
	HttpTimeout        time.Duration `mapstructure:"http_timeout" validate:"required"`
}

type ovhChallenge struct {
	*legoOVH.DNSProvider
	id string
}

func (g *ovhChallenge) ID() string {
	return fmt.Sprintf("%s-%s", KeyDnsOVH, g.id)
}

func (g *ovhChallenge) Type() string {
	return acme.TypeDNS01
}

func CreateOvh(_ *context.ServerContext, id string, cfg map[string]interface{}) (acme.Challenge, error) {
	config := legoOVH.NewDefaultConfig()
	instanceConfig := ConfigOvh{
		Endpoint:           "ovh-eu",
		PropagationTimeout: config.PropagationTimeout,
		PollingInterval:    config.PollingInterval,
		HttpTimeout:        config.HTTPClient.Timeout,
	}
	err := mapstructure.Decode(cfg, &instanceConfig)
	if err != nil {
		return nil, err
	}

	//validate := validator.New()
	//err = validate.Struct(instanceConfig)
	//if err != nil {
	//	return nil, err
	//}

	config.AccessToken = instanceConfig.AccessToken
	config.ApplicationKey = instanceConfig.ApplicationKey
	config.ApplicationSecret = instanceConfig.ApplicationSecret

	if instanceConfig.ClientID != "" || instanceConfig.ClientSecret != "" {
		config.OAuth2Config = &legoOVH.OAuth2Config{
			ClientID:     instanceConfig.ClientID,
			ClientSecret: instanceConfig.ClientSecret,
		}
	}

	config.ConsumerKey = instanceConfig.ConsumerKey
	config.APIEndpoint = instanceConfig.Endpoint
	provider, err := legoOVH.NewDNSProviderConfig(config)
	return &ovhChallenge{id: id, DNSProvider: provider}, err
}
