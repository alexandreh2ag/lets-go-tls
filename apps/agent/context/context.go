package context

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/state"
	goCacheLib "github.com/eko/gocache/lib/v4/cache"
	gocacheStore "github.com/eko/gocache/store/go_cache/v4"
	go_cache "github.com/patrickmn/go-cache"
	"io"
	"time"
)

var _ context.Context = &AgentContext{}

type AgentContext struct {
	*context.BaseContext

	Config       *config.Config
	Requesters   types.Requesters
	Storages     certificate.Storages
	StateStorage state.Storage
	Cache        types.Cache
}

func DefaultContext() *AgentContext {
	cfg := config.DefaultConfig()

	return &AgentContext{
		BaseContext: context.DefaultContext(),
		Config:      &cfg,
	}
}

func TestContext(logBuffer io.Writer) *AgentContext {
	if logBuffer == nil {
		logBuffer = io.Discard
	}
	cfg := config.DefaultConfig()

	cacheClient := go_cache.New(time.Second*2, time.Millisecond*200)
	cacheStore := gocacheStore.NewGoCache(cacheClient)
	cacheManager := goCacheLib.New[string](cacheStore)
	return &AgentContext{
		BaseContext: context.TestContext(logBuffer),
		Config:      &cfg,
		Cache:       cacheManager,
	}
}
