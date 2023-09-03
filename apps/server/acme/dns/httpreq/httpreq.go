package httpreq

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/providers/dns/httpreq"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"time"
)

const KeyDnsHttpReq = "httpreq"

type ConfigHttpReq struct {
	Endpoint           string        `mapstructure:"endpoint" validate:"required"`
	Mode               string        `mapstructure:"mode"`
	Username           string        `mapstructure:"username"`
	Password           string        `mapstructure:"password"`
	PropagationTimeout time.Duration `mapstructure:"propagation_timeout" validate:"required"`
	PollingInterval    time.Duration `mapstructure:"polling_interval" validate:"required"`
	HttpTimeout        time.Duration `mapstructure:"http_timeout" validate:"required"`
}

type httpReq struct {
	*httpreq.DNSProvider
	id string
}

func (g *httpReq) ID() string {
	return fmt.Sprintf("%s-%s", KeyDnsHttpReq, g.id)
}

func (g *httpReq) Type() string {
	return acme.TypeDNS01
}

func CreateHttpReq(_ *context.ServerContext, id string, cfg map[string]interface{}) (acme.Challenge, error) {
	config := httpreq.NewDefaultConfig()
	instanceConfig := ConfigHttpReq{
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

	endpoint, err := url.Parse(instanceConfig.Endpoint)
	if err != nil {
		return nil, err
	}

	config.Endpoint = endpoint
	config.Mode = instanceConfig.Mode
	config.Username = instanceConfig.Username
	config.Password = instanceConfig.Password
	config.PollingInterval = instanceConfig.PollingInterval
	config.PropagationTimeout = instanceConfig.PropagationTimeout
	config.HTTPClient.Timeout = instanceConfig.HttpTimeout

	provider, err := httpreq.NewDNSProviderConfig(config)
	return &httpReq{id: id, DNSProvider: provider}, err
}
