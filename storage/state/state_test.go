package state

import (
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateStorage_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	want := &fs{fs: ctx.Fs, logger: ctx.GetLogger(), cfg: ConfigFs{Path: "/app/acme.json"}, checksum: appFs.NewChecksum(ctx.Fs)}
	cfg := config.StateConfig{Type: FsKey, Config: map[string]interface{}{"path": "/app/acme.json"}}
	got, err := CreateStorage(ctx, cfg)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCreateStorage_Fail(t *testing.T) {
	ctx := context.TestContext(nil)
	cfg := config.StateConfig{Type: "wrong", Config: map[string]interface{}{}}
	got, err := CreateStorage(ctx, cfg)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config state storage type 'wrong' does not exist")
}
