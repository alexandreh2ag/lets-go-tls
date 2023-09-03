package context

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestDefaultContext(t *testing.T) {

	cfg := config.DefaultConfig()

	want := &ServerContext{
		Config: &cfg,
	}
	got := DefaultContext()
	got.BaseContext = want.BaseContext
	assert.Equal(t, want, got)
}

func TestTestContext(t *testing.T) {

	cfg := config.DefaultConfig()
	want := &ServerContext{
		Config: &cfg,
	}
	got := TestContext(nil)
	got.BaseContext = want.BaseContext
	got.Cache = nil
	assert.Equal(t, want, got)
}

func TestTestContext_WithLogBuffer(t *testing.T) {

	cfg := config.DefaultConfig()
	want := &ServerContext{
		Config: &cfg,
	}
	got := TestContext(io.Discard)
	got.BaseContext = want.BaseContext
	got.Cache = nil
	assert.Equal(t, want, got)
}
