package context

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/state"
	goCacheLib "github.com/eko/gocache/lib/v4/cache"
	gocacheStore "github.com/eko/gocache/store/go_cache/v4"
	go_cache "github.com/patrickmn/go-cache"
	"io"
	"time"
)

var _ context.Context = &ServerContext{}

type ServerContext struct {
	*context.BaseContext

	Config       *config.Config
	Requesters   types.Requesters
	StateStorage state.Storage
	Cache        types.Cache
}

func DefaultContext() *ServerContext {
	cfg := config.DefaultConfig()

	return &ServerContext{
		BaseContext: context.DefaultContext(),
		Config:      &cfg,
	}
}

func TestContext(logBuffer io.Writer) *ServerContext {
	if logBuffer == nil {
		logBuffer = io.Discard
	}
	cfg := config.DefaultConfig()

	cacheClient := go_cache.New(time.Second*2, time.Millisecond*200)
	cacheStore := gocacheStore.NewGoCache(cacheClient)
	cacheManager := goCacheLib.New[string](cacheStore)
	return &ServerContext{
		BaseContext: context.TestContext(logBuffer),
		Config:      &cfg,
		Cache:       cacheManager,
	}
}
