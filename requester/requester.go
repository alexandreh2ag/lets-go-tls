package requester

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
)

var TypeRequesterMapping = map[string]CreateRequesterFn{}

type CreateRequesterFn func(ctx context.Context, config config.RequesterConfig) (types.Requester, error)

func CreateRequesters(ctx context.Context, requestersConfig []config.RequesterConfig) (types.Requesters, error) {
	instances := types.Requesters{}
	var err error = nil
	ctx.GetLogger().Info("Create requesters")
	for _, cfg := range requestersConfig {
		ctx.GetLogger().Debug(fmt.Sprintf("Create requester %s (type: %s)", cfg.Id, cfg.Type))
		instances[cfg.Id], err = createRequester(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}
	return instances, err
}

func createRequester(ctx context.Context, cfg config.RequesterConfig) (types.Requester, error) {
	if fn, ok := TypeRequesterMapping[cfg.Type]; ok {
		return fn(ctx, cfg)
	}
	return nil, fmt.Errorf("config requester type '%s' does not exist", cfg.Type)
}
