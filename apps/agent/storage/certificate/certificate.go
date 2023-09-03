package certificate

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
)

var TypeStorageMapping = map[string]CreateStorageFn{}

type CreateStorageFn func(ctx *context.AgentContext, config config.StorageConfig) (certificate.Storage, error)

func CreateCertificateStorages(ctx *context.AgentContext) (certificate.Storages, error) {
	instances := certificate.Storages{}
	var err error = nil
	ctx.Logger.Info("Create certificate storages")
	for _, cfgStorage := range ctx.Config.Storages {
		ctx.Logger.Debug(fmt.Sprintf("Create certificate storage %s (type: %s)", cfgStorage.Id, cfgStorage.Type))
		instances[cfgStorage.Id], err = createStorage(ctx, cfgStorage)
		if err != nil {
			return nil, err
		}
	}
	return instances, err
}

func createStorage(ctx *context.AgentContext, cfgProvider config.StorageConfig) (certificate.Storage, error) {
	if fn, ok := TypeStorageMapping[cfgProvider.Type]; ok {
		return fn(ctx, cfgProvider)
	}
	return nil, fmt.Errorf("config certificate storage type '%s' does not exist", cfgProvider.Type)
}
