package certificate

import (
	"errors"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/hook"
	mockAfero "github.com/alexandreh2ag/lets-go-tls/mocks/afero"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"os"
	"testing"
)

func Test_nginx_ID(t *testing.T) {
	n := nginx{id: "foo"}
	assert.Equalf(t, "foo", n.ID(), "ID()")
}

func Test_createNginxStorage(t *testing.T) {
	ctx := context.TestContext(nil)
	uid := os.Getuid()
	gid := os.Getgid()
	cfgFsFile := ConfigFs{Path: "/etc/nginx"}
	cfg := ConfigNginx{ConfigFs: cfgFsFile, NginxCfgPath: "/etc/nginx/nginx.conf"}

	fsStorage := &fs{id: "foo", fs: ctx.Fs, cfg: cfgFsFile, checksum: appFs.NewChecksum(ctx.Fs), uid: uid, gid: gid}
	storage := &nginx{id: "foo", fsStorage: fsStorage, cfg: cfg, uid: uid, gid: gid, logger: ctx.Logger}

	tests := []struct {
		name        string
		cfg         config.StorageConfig
		want        certificate.Storage
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg: config.StorageConfig{
				Id:     "foo",
				Config: map[string]interface{}{"nginx_cfg_path": "/etc/nginx/nginx.conf"},
			},
			want: storage,
		},
		{
			name: "FailDecodeCfg",
			cfg: config.StorageConfig{
				Id:     "foo",
				Config: map[string]interface{}{"nginx_cfg_path": []string{"foo"}},
			},
			want:        storage,
			wantErr:     true,
			errContains: "'nginx_cfg_path' expected type 'string', got unconvertible type '[]string'",
		},
		{
			name: "FailValidateCfg",
			cfg: config.StorageConfig{
				Id:     "foo",
				Config: map[string]interface{}{},
			},
			want:        storage,
			wantErr:     true,
			errContains: "Error:Field validation for 'NginxCfgPath' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createNginxStorage(ctx, tt.cfg)

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

func Test_nginx_Save_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	identifier1 := "example.com-0"
	identifier2 := "foo.example.com-0"
	certificates := types.Certificates{
		{Identifier: identifier1, Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: identifier2, Domains: types.Domains{"foo.example.com", "bar.example.com"}, Key: []byte("key2"), Certificate: []byte("certificate2")},
	}
	chanHook := make(chan *hook.Hook)
	postHook := &hook.Hook{Cmd: "echo 1"}
	go func() {
		for {
			select {
			case h := <-chanHook:
				assert.Equal(t, postHook, h)
			}
		}
	}()
	cfg := ConfigNginx{
		NginxCfgPath: "../../../../nginx/fixtures/nginx_valid/nginx.conf",
		ConfigFs: ConfigFs{
			Path:            "/app",
			SpecificDomains: []ConfigSpecificDomain{},
			PostHook:        postHook,
		},
	}
	checksum := appFs.NewChecksum(ctx.Fs)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: ctx.Fs, checksum: checksum}
	storage := &nginx{
		cfg:       cfg,
		fsStorage: fsStorage,
		logger:    ctx.Logger,
	}
	errs := storage.Save(certificates, chanHook)

	assert.Len(t, errs, 0)

	contentKey, err := afero.ReadFile(ctx.Fs, "/etc/ssl/example.com.key")
	assert.NoError(t, err)
	assert.Equal(t, "key", string(contentKey))

	contentCrt, err := afero.ReadFile(ctx.Fs, "/etc/ssl/example.com.crt")
	assert.NoError(t, err)
	assert.Equal(t, "certificate", string(contentCrt))

	contentKey, err = afero.ReadFile(ctx.Fs, "/etc/ssl/foo.example.com.key")
	assert.NoError(t, err)
	assert.Equal(t, "key2", string(contentKey))

	contentCrt, err = afero.ReadFile(ctx.Fs, "/etc/ssl/foo.example.com.crt")
	assert.NoError(t, err)
	assert.Equal(t, "certificate2", string(contentCrt))

}

func Test_nginx_Save_FailedParseConfigNginx(t *testing.T) {
	ctx := context.TestContext(nil)
	identifier1 := "example.com-0"
	identifier2 := "foo.example.com-0"
	certificates := types.Certificates{
		{Identifier: identifier1, Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: identifier2, Domains: types.Domains{"foo.example.com", "bar.example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	chanHook := make(chan *hook.Hook)

	cfg := ConfigNginx{
		NginxCfgPath: "./wrong/nginx.conf",
		ConfigFs: ConfigFs{
			Path:            "/app",
			SpecificDomains: []ConfigSpecificDomain{},
		},
	}
	checksum := appFs.NewChecksum(ctx.Fs)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: ctx.Fs, checksum: checksum}
	storage := &nginx{
		cfg:       cfg,
		fsStorage: fsStorage,
		logger:    ctx.Logger,
	}
	errs := storage.Save(certificates, chanHook)

	assert.Len(t, errs, 1)
}

func Test_nginx_Save_FailedWriteKey(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	gomock.InOrder(
		// Key
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("error")),
	)
	identifier1 := "example.com-0"
	certificates := types.Certificates{
		{Identifier: identifier1, Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	chanHook := make(chan *hook.Hook)

	cfg := ConfigNginx{
		NginxCfgPath: "../../../../nginx/fixtures/nginx_valid/nginx.conf",
		ConfigFs: ConfigFs{
			Path:            "/app",
			SpecificDomains: []ConfigSpecificDomain{},
		},
	}
	checksum := appFs.NewChecksum(fsMock)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: fsMock, checksum: checksum}
	storage := &nginx{
		cfg:       cfg,
		fsStorage: fsStorage,
		logger:    ctx.Logger,
	}
	errs := storage.Save(certificates, chanHook)

	assert.Len(t, errs, 1)
}

func Test_nginx_Save_FailedWriteCert(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)

	file, _ := ctx.Fs.Create("/app/test.txt")
	gomock.InOrder(
		// Key
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(file, nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),

		// Certificate
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("error")),
	)
	identifier1 := "example.com-0"
	certificates := types.Certificates{
		{Identifier: identifier1, Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	chanHook := make(chan *hook.Hook)

	cfg := ConfigNginx{
		NginxCfgPath: "../../../../nginx/fixtures/nginx_valid/nginx.conf",
		ConfigFs: ConfigFs{
			Path:            "/app",
			SpecificDomains: []ConfigSpecificDomain{},
		},
	}
	checksum := appFs.NewChecksum(fsMock)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: fsMock, checksum: checksum}
	storage := &nginx{
		cfg:       cfg,
		fsStorage: fsStorage,
		logger:    ctx.Logger,
	}
	errs := storage.Save(certificates, chanHook)

	assert.Len(t, errs, 1)
}

func Test_nginx_Delete(t *testing.T) {
	chanHook := make(chan *hook.Hook)
	certificates := types.Certificates{
		{Identifier: "example.com-0", Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	n := nginx{}
	assert.Equalf(t, []error{}, n.Delete(certificates, chanHook), "Delete(%v, %v)", certificates, chanHook)
}
