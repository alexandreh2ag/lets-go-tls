package certificate

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_createStorage_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	want := &fs{id: "foo", fs: ctx.Fs, cfg: ConfigFs{Path: "/app"}, checksum: appFs.NewChecksum(ctx.Fs)}
	cfg := config.StorageConfig{
		Id:     "foo",
		Type:   "fs",
		Config: map[string]interface{}{"path": "/app"},
	}
	got, err := createStorage(ctx, cfg)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_createStorage_Fail(t *testing.T) {
	ctx := context.TestContext(nil)
	cfg := config.StorageConfig{
		Id:     "foo",
		Type:   "wrong",
		Config: map[string]interface{}{},
	}
	got, err := createStorage(ctx, cfg)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config certificate storage type 'wrong' does not exist")
}

func TestCreateStorages_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.Config.Storages = []config.StorageConfig{
		{
			Id:     "foo",
			Type:   "fs",
			Config: map[string]interface{}{"path": "/app"},
		},
	}
	staticP := &fs{id: "foo", fs: ctx.Fs, cfg: ConfigFs{Path: "/app"}, checksum: appFs.NewChecksum(ctx.Fs)}
	want := certificate.Storages{"foo": staticP}
	got, err := CreateCertificateStorages(ctx)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCreateCertificateStorages_Fail(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.Config.Storages = []config.StorageConfig{
		{
			Id:     "foo",
			Type:   "wrong",
			Config: map[string]interface{}{},
		},
	}

	got, err := CreateCertificateStorages(ctx)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config certificate storage type 'wrong' does not exist")
}
