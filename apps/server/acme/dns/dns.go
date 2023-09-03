package dns

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns/gandiv5"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns/httpreq"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns/ovh"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
)

var TypeDnsProviderMapping = map[string]CreateDnsChallengeFn{
	gandiv5.KeyDnsGandiV5: gandiv5.CreateGandiV5,
	httpreq.KeyDnsHttpReq: httpreq.CreateHttpReq,
	ovh.KeyDnsOVH:         ovh.CreateOvh,
}

type CreateDnsChallengeFn func(ctx *context.ServerContext, id string, config map[string]interface{}) (acme.Challenge, error)

func CreateDnsChallenge(ctx *context.ServerContext, id string, cfg config.ResolverConfig) (acme.Challenge, error) {
	if fn, ok := TypeDnsProviderMapping[cfg.Type]; ok {
		return fn(ctx, id, cfg.Config)
	}
	return nil, fmt.Errorf("config dns challenge id '%s' (type %s) does not exist", id, cfg.Type)
}
