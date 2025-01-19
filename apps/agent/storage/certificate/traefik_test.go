package certificate

import (
	"errors"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/hook"
	mockAfero "github.com/alexandreh2ag/lets-go-tls/mocks/afero"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"os"
	"path/filepath"
	"testing"
)

func Test_traefik_ID(t *testing.T) {
	f := &traefik{id: "foo"}
	assert.Equal(t, "foo", f.ID())
}

func Test_createTraefikStorage(t *testing.T) {
	ctx := context.TestContext(nil)
	uid := os.Getuid()
	gid := os.Getgid()
	storage := &traefik{id: "foo", fs: ctx.Fs, cfg: ConfigTraefik{Path: "/app"}, checksum: appFs.NewChecksum(ctx.Fs), uid: uid, gid: gid}
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
				Config: map[string]interface{}{"path": "/app"},
			},
			want: storage,
		},
		{
			name: "FailDecodeCfg",
			cfg: config.StorageConfig{
				Id:     "foo",
				Config: map[string]interface{}{"path": []string{"foo"}},
			},
			want:        storage,
			wantErr:     true,
			errContains: "'path' expected type 'string', got unconvertible type '[]string'",
		},
		{
			name: "FailValidateCfg",
			cfg: config.StorageConfig{
				Id:     "foo",
				Config: map[string]interface{}{"path": ""},
			},
			want:        storage,
			wantErr:     true,
			errContains: "Key: 'ConfigTraefik.Path' Error:Field validation for 'Path' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createTraefikStorage(ctx, tt.cfg)

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

func Test_traefik_Save_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	identifier := "example.com"
	certificates := types.Certificates{
		{Identifier: identifier, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	storage := &traefik{fs: ctx.Fs, cfg: ConfigTraefik{Path: "/app"}, checksum: appFs.NewChecksum(ctx.Fs)}
	errs := storage.Save(certificates, make(chan<- *hook.Hook))
	assert.Len(t, errs, 0)
	contentKey, err := afero.ReadFile(ctx.Fs, filepath.Join(storage.cfg.Path, identifier+".yml"))
	assert.NoError(t, err)
	assert.Equal(t, "tls:\n  certificates:\n  - certFile: certificate\n    keyFile: key\n", string(contentKey))

}

func Test_traefik_Save_FailCreateDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	fsMock.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Times(1).Return(errors.New("error"))
	certificates := types.Certificates{}
	storage := &traefik{fs: fsMock, cfg: ConfigTraefik{Path: "/app"}}
	errs := storage.Save(certificates, make(chan<- *hook.Hook))
	assert.Len(t, errs, 1)
}

func Test_traefik_Save_FailWriteCertFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	certificates := types.Certificates{
		{Identifier: "example.com", Key: []byte("key"), Certificate: []byte("certificate")},
	}
	gomock.InOrder(
		fsMock.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Times(1).Return(nil),
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("error")),
	)
	storage := &traefik{fs: fsMock, cfg: ConfigTraefik{Path: "/app"}, checksum: appFs.NewChecksum(fsMock)}
	errs := storage.Save(certificates, make(chan<- *hook.Hook))
	assert.Len(t, errs, 1)
}

func Test_traefik_GetFilePath_Success(t1 *testing.T) {
	want := "/app/example.com-0.yml"
	cert := &types.Certificate{Identifier: "example.com-0"}
	t := &traefik{cfg: ConfigTraefik{Path: "/app"}}
	assert.Equal(t1, want, t.GetFilePath(cert))
}

func Test_traefik_GetFilePath_SuccessWithPrefix(t1 *testing.T) {
	want := "/app/ssl.example.com-0.yml"
	cert := &types.Certificate{Identifier: "example.com-0"}
	t := &traefik{cfg: ConfigTraefik{Path: "/app", PrefixFilename: "ssl."}}
	assert.Equal(t1, want, t.GetFilePath(cert))
}

func Test_traefik_Delete(t1 *testing.T) {

	ctrl := gomock.NewController(t1)
	fsMock := mockAfero.NewMockFs(ctrl)

	gomock.InOrder(
		fsMock.EXPECT().Remove(gomock.Any()).Times(1).Return(nil),
		fsMock.EXPECT().Remove(gomock.Any()).Times(1).Return(errors.New("error")),
	)
	certificates := types.Certificates{
		{Identifier: "example.com-0", Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: "example.com-1", Key: []byte("key"), Certificate: []byte("certificate")},
	}
	want := []error{errors.New("error")}
	t := traefik{
		fs:  fsMock,
		cfg: ConfigTraefik{Path: "/app"},
	}
	assert.Equalf(t1, want, t.Delete(certificates, make(chan<- *hook.Hook)), "Delete(%v)", certificates)
}

func Test_traefik_WriteCertFile(t1 *testing.T) {
	cert := &types.Certificate{Identifier: "example.com-0", Certificate: []byte("certificate"), Key: []byte("key")}
	ctrl := gomock.NewController(t1)

	tests := []struct {
		name    string
		mockFn  func(fs *mockAfero.MockFs)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			mockFn: func(fs *mockAfero.MockFs) {
				fs.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error"))
				fs.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(mem.NewFileHandle(mem.CreateFile("/app/file")), nil)
				fs.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
			wantErr: assert.NoError,
		},
		{
			name: "FailWrite",
			mockFn: func(fs *mockAfero.MockFs) {
				fs.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error"))
				fs.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("fail write"))
			},
			wantErr: assert.Error,
		},

		{
			name: "FailChown",
			mockFn: func(fs *mockAfero.MockFs) {
				fs.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error"))
				fs.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(mem.NewFileHandle(mem.CreateFile("/app/file")), nil)
				fs.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail chown"))
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			fsMock := mockAfero.NewMockFs(ctrl)
			tt.mockFn(fsMock)
			checksum := appFs.NewChecksum(fsMock)
			t := traefik{
				id:       "traefik",
				fs:       fsMock,
				checksum: checksum,
				cfg:      ConfigTraefik{Path: "/app"},
			}
			tt.wantErr(t1, t.WriteCertFile(cert), fmt.Sprintf("WriteCertFile(%v)", cert))
		})
	}
}
