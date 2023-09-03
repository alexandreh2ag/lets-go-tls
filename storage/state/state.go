package state

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/state"
)

var TypeStorageMapping = map[string]CreateStorageFn{}

type CreateStorageFn func(ctx context.Context, cfg config.StateConfig) (state.Storage, error)

func CreateStorage(ctx context.Context, cfg config.StateConfig) (state.Storage, error) {
	if fn, ok := TypeStorageMapping[cfg.Type]; ok {
		return fn(ctx, cfg)
	}
	return nil, fmt.Errorf("config state storage type '%s' does not exist", cfg.Type)
}
