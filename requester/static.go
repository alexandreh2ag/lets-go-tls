package requester

import (
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"slices"
)

const (
	StaticKey = "static"
)

func init() {
	TypeRequesterMapping[StaticKey] = createStaticProvider
}

var _ types.Requester = &static{}

type ConfigStatic struct {
	ListDomains [][]string `mapstructure:"domains" validate:"required,min=1"`
}

type static struct {
	id             string
	domainRequests []*types.DomainRequest
}

func (f *static) ID() string {
	return f.id
}

func (f *static) Fetch() ([]*types.DomainRequest, error) {
	return f.domainRequests, nil
}

func createStaticProvider(_ context.Context, cfg config.RequesterConfig) (types.Requester, error) {
	instanceConfig := ConfigStatic{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	instance := &static{id: cfg.Id}
	for _, ListDomain := range instanceConfig.ListDomains {
		for _, domain := range ListDomain {
			domainRequest := &types.DomainRequest{Requester: instance, Domains: types.Domains{types.Domain(domain)}}
			instance.domainRequests = append(instance.domainRequests, domainRequest)
		}
	}

	instance.domainRequests = slices.DeleteFunc(instance.domainRequests, func(item *types.DomainRequest) bool {
		if item.IsIP() {
			return true
		}
		return false
	})

	return instance, nil
}
