package state

import (
	"encoding/json"
	"github.com/alexandreh2ag/lets-go-tls/config"
	appCtx "github.com/alexandreh2ag/lets-go-tls/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/state"
	legoAcme "github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/registration"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"io"
	"path"
	"testing"
)

func Test_fs_Load_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	basePath := "/app"
	stateStorage := &fs{fs: ctx.Fs, cfg: ConfigFs{Path: path.Join(basePath, "acme.json")}, checksum: appFs.NewChecksum(ctx.Fs)}

	want := &types.State{
		Account:      &acme.Account{Email: "dev@foo.com", Registration: &registration.Resource{Body: legoAcme.Account{Status: "valid"}, URI: "https://uri.com"}, Key: []byte("privatekey")},
		Certificates: types.Certificates{{Domains: types.Domains{"foo.com", "bar.com"}, Key: []byte("key"), Certificate: []byte("certificate")}},
	}
	data, _ := json.Marshal(want)
	_ = ctx.Fs.Mkdir(basePath, 0775)
	_ = afero.WriteFile(ctx.Fs, stateStorage.cfg.Path, data, 0644)

	got, err := stateStorage.Load()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_fs_Load_SuccessPathNotExist(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	basePath := "/app"
	stateStorage := &fs{fs: ctx.Fs, cfg: ConfigFs{Path: path.Join(basePath, "acme.json")}}

	want := &types.State{Account: &acme.Account{}}

	got, err := stateStorage.Load()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_fs_Load_FailMarshalFile(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	basePath := "/app"
	stateStorage := &fs{fs: ctx.Fs, cfg: ConfigFs{Path: path.Join(basePath, "acme.json")}}

	_ = ctx.Fs.Mkdir(basePath, 0775)
	_ = afero.WriteFile(ctx.Fs, stateStorage.cfg.Path, []byte("[}"), 0644)
	got, err := stateStorage.Load()
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse json /app/acme.json: invalid character '}' looking for beginning of value")
}

func Test_fs_Save_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	basePath := "/app"
	stateStorage := &fs{fs: ctx.Fs, cfg: ConfigFs{Path: path.Join(basePath, "acme.json")}, checksum: appFs.NewChecksum(ctx.Fs)}
	s := &types.State{
		Account:      &acme.Account{Email: "dev@foo.com", Registration: &registration.Resource{Body: legoAcme.Account{Status: "valid"}, URI: "https://uri.com"}, Key: []byte("privatekey")},
		Certificates: types.Certificates{{Domains: types.Domains{"foo.com", "bar.com"}, Key: []byte("key"), Certificate: []byte("certificate")}},
	}
	data, _ := json.Marshal(s)
	err := stateStorage.Save(s)
	assert.NoError(t, err)
	fileExist, _ := afero.Exists(ctx.Fs, stateStorage.cfg.Path)
	file, _ := ctx.Fs.Open(stateStorage.cfg.Path)
	fileData, _ := io.ReadAll(file)
	assert.True(t, fileExist)
	assert.Equal(t, data, fileData)
}

func Test_fs_Type(t *testing.T) {
	stateStorage := &fs{}
	assert.Equal(t, FsKey, stateStorage.Type())
}

func Test_createFsStorage(t *testing.T) {
	ctx := appCtx.TestContext(nil)

	tests := []struct {
		name        string
		cfg         config.StateConfig
		want        state.Storage
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg: config.StateConfig{
				Type: FsKey,
				Config: map[string]interface{}{
					"path": "/app/acme.jsom",
				},
			},
			want: &fs{fs: ctx.Fs, cfg: ConfigFs{Path: "/app/acme.jsom"}, checksum: appFs.NewChecksum(ctx.Fs)},
		},
		{
			name: "FailDecodeCfg",
			cfg: config.StateConfig{
				Type: FsKey,
				Config: map[string]interface{}{
					"path": []string{},
				},
			},
			wantErr:     true,
			errContains: "path' expected type 'string', got unconvertible type '[]string",
		},
		{
			name: "FailValidateCfg",
			cfg: config.StateConfig{
				Type: FsKey,
				Config: map[string]interface{}{
					"path": "",
				},
			},
			wantErr:     true,
			errContains: "ConfigFs.Path' Error:Field validation for 'Path' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createFsStorage(ctx, tt.cfg)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
