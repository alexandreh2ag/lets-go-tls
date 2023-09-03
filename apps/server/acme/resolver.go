package acme

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/http"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	legoLog "github.com/go-acme/lego/v4/log"
	"io"
	"log"
)

var (
	httpProvider *http.ChallengeHTTP
)

func CreateResolvers(ctx *context.ServerContext, account *acme.Account) (types.Resolvers, error) {
	var err error = nil
	instances := types.Resolvers{}

	legoLog.Logger = log.New(io.Discard, "", log.LstdFlags)
	configAcme := lego.NewConfig(account)
	configAcme.CADirURL = ctx.Config.Acme.CAServer
	configAcme.Certificate.KeyType = certcrypto.RSA4096

	ctx.Config.Acme.Resolvers[types.DefaultKey] = config.ResolverConfig{
		Type:    acme.TypeHTTP01,
		Filters: []string{"*"},
	}

	ctx.Logger.Info("Create acme resolvers")
	for id, cfgResolver := range ctx.Config.Acme.Resolvers {
		ctx.Logger.Debug(fmt.Sprintf("Create acme resolver %s ", id))
		resolver, errCreateResolver := createResolver(ctx, id, cfgResolver, configAcme)
		if errCreateResolver != nil {
			return nil, errCreateResolver
		}

		instances[id] = resolver
	}
	return instances, err
}

func createResolver(ctx *context.ServerContext, id string, cfg config.ResolverConfig, cfgAcme *lego.Config) (types.Resolver, error) {
	var provider acme.Challenge
	client, err := lego.NewClient(cfgAcme)
	if err != nil {
		return nil, fmt.Errorf("failed to init acme client for resolver %s: %v", id, err)
	}
	resolver := &ResolverAcme{Id: id, Filters: cfg.Filters, Client: client}

	if cfg.Type == acme.TypeHTTP01 {
		provider = GetHTTPProvider(ctx)
		err = client.Challenge.SetHTTP01Provider(provider)
	} else {
		provider, err = dns.CreateDnsChallenge(ctx, id, cfg)
		if err != nil {
			return nil, err
		}
		err = client.Challenge.SetDNS01Provider(provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to register provider for resovler %s: %v", id, err)
	}

	resolver.Challenge = provider

	return resolver, nil
}

func GetHTTPProvider(ctx *context.ServerContext) *http.ChallengeHTTP {
	if httpProvider != nil {
		return httpProvider
	}
	httpProvider = http.NewChallenge(ctx.Logger, ctx.Cache)
	return httpProvider
}
