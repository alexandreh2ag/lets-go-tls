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
	"path/filepath"
	"testing"
)

func Test_haproxy_ID(t *testing.T) {
	f := &haproxy{id: "foo"}
	assert.Equal(t, "foo", f.ID())
}

func Test_createHaproxyStorage(t *testing.T) {
	ctx := context.TestContext(nil)
	uid := os.Getuid()
	gid := os.Getgid()
	cfgFsFile := ConfigFs{Path: "/app", AddPem: true}
	cfg := ConfigHaproxy{ConfigFs: cfgFsFile, CrtListPath: "/app/crt-list.txt"}
	cfgWithCrtList := ConfigHaproxy{ConfigFs: cfgFsFile, CrtListPath: "/crt-list.txt"}

	fsStorage := &fs{id: "foo", fs: ctx.Fs, cfg: cfgFsFile, checksum: appFs.NewChecksum(ctx.Fs), uid: uid, gid: gid}
	storage := &haproxy{id: "foo", fs: ctx.Fs, fsStorage: fsStorage, cfg: cfg, checksum: appFs.NewChecksum(ctx.Fs), uid: uid, gid: gid}
	storageWithCrtList := &haproxy{id: "foo", fs: ctx.Fs, fsStorage: fsStorage, cfg: cfgWithCrtList, checksum: appFs.NewChecksum(ctx.Fs), uid: uid, gid: gid}

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
			name: "SuccessWithCrtListPath",
			cfg: config.StorageConfig{
				Id:     "foo",
				Config: map[string]interface{}{"path": "/app", "crt_list_path": "/crt-list.txt"},
			},
			want: storageWithCrtList,
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
				Config: map[string]interface{}{},
			},
			want:        storage,
			wantErr:     true,
			errContains: "Error:Field validation for 'Path' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createHaproxyStorage(ctx, tt.cfg)

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

func Test_haproxy_Save_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	identifier1 := "example.com-0"
	identifier2 := "foo.com-0"
	identifierCustom := "foo-custom"
	certificates := types.Certificates{
		{Identifier: identifier1, Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: identifier2, Domains: types.Domains{"foo.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
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
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			Path:   "/app",
			AddPem: true,
			SpecificDomains: []ConfigSpecificDomain{
				{Identifier: identifierCustom, Domains: types.Domains{"foo.com"}},
			},
			PostHook: postHook,
		},
	}
	checksum := appFs.NewChecksum(ctx.Fs)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: ctx.Fs, checksum: checksum}
	storage := &haproxy{
		fs:        ctx.Fs,
		cfg:       cfg,
		fsStorage: fsStorage,

		checksum: checksum,
	}
	errs := storage.Save(certificates, chanHook)

	assert.Len(t, errs, 0)
	contentKey, err := afero.ReadFile(ctx.Fs, filepath.Join(storage.cfg.Path, identifier1+".key"))
	assert.NoError(t, err)
	assert.Equal(t, "key", string(contentKey))

	contentCrt, err := afero.ReadFile(ctx.Fs, filepath.Join(storage.cfg.Path, identifier1+".crt"))
	assert.NoError(t, err)
	assert.Equal(t, "certificate", string(contentCrt))

	contentKey, err = afero.ReadFile(ctx.Fs, filepath.Join(storage.cfg.Path, identifierCustom+".key"))
	assert.NoError(t, err)
	assert.Equal(t, "key", string(contentKey))

	contentCrt, err = afero.ReadFile(ctx.Fs, filepath.Join(storage.cfg.Path, identifierCustom+".crt"))
	assert.NoError(t, err)
	assert.Equal(t, "certificate", string(contentCrt))

	crtList, errCrtList := afero.ReadFile(ctx.Fs, cfg.CrtListPath)
	assert.NoError(t, errCrtList)
	assert.Equal(t, "\n/app/example.com-0.pem example.com\n/app/foo-custom.pem foo.com", string(crtList))
}

func Test_haproxy_Save_Success_FailedGenerateCrtList(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	certificates := types.Certificates{
		{Identifier: "example.com", Key: []byte("key"), Certificate: []byte("certificate")},
	}
	file, _ := ctx.Fs.Create("/app/test.txt")
	file2, _ := ctx.Fs.Create("/app/test.txt")
	file3, _ := ctx.Fs.Create("/app/test.txt")
	file4, _ := ctx.Fs.Create("/app/test.txt")
	gomock.InOrder(
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, errors.New("fail")),
		fsMock.EXPECT().Mkdir(gomock.Any(), gomock.Any()).Times(1).Return(nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),
		// key file
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(file, nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),
		// cert file
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(file2, nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),
		// pem file
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(file3, nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),

		// crt list
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, errors.New("fail")),
		fsMock.EXPECT().Mkdir(gomock.Any(), gomock.Any()).Times(1).Return(nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),

		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(file4, nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail chown")),
	)
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			AddPem: true,
			Path:   "/app",
		},
	}
	checksum := appFs.NewChecksum(fsMock)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: fsMock, checksum: checksum}
	storage := &haproxy{
		fs:        fsMock,
		cfg:       cfg,
		fsStorage: fsStorage,

		checksum: checksum,
	}
	errs := storage.Save(certificates, make(chan *hook.Hook))
	assert.Len(t, errs, 1)
}

func Test_fs_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)

	gomock.InOrder(
		//example.com-0
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, nil),
		fsMock.EXPECT().Remove(gomock.Any()).Times(1).Return(nil),
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, nil),
		fsMock.EXPECT().Remove(gomock.Any()).Times(1).Return(nil),
		//example.com-1
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, nil),
		fsMock.EXPECT().Remove(gomock.Any()).Times(1).Return(errors.New("error")),
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, nil),
		fsMock.EXPECT().Remove(gomock.Any()).Times(1).Return(errors.New("error")),
		//example.com-2
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, errors.New("error")),
	)
	certificates := types.Certificates{
		{Identifier: "example.com-0", Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: "example.com-1", Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: "example.com-2", Key: []byte("key"), Certificate: []byte("certificate")},
	}
	want := []error{errors.New("error"), errors.New("error")}
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
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			Path:   "/app",
			AddPem: true,
		},
	}
	checksum := appFs.NewChecksum(fsMock)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: fsMock, checksum: checksum}
	storage := &haproxy{
		fs:        fsMock,
		cfg:       cfg,
		fsStorage: fsStorage,

		checksum: checksum,
	}
	assert.Equalf(t, want, storage.Delete(certificates, chanHook), "Delete(%v)", certificates)
}

func Test_haproxy_generateCrtListContent_Success(t *testing.T) {
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			Path:   "/app",
			AddPem: true,
		},
	}
	fsStorage := &fs{cfg: cfg.ConfigFs}
	h := haproxy{
		fsStorage: fsStorage,
		cfg:       cfg,
	}
	certificates := types.Certificates{
		{Identifier: "example.com-0", Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: "example2.com-0", Domains: types.Domains{"example2.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	want := "\n/app/example.com-0.pem example.com\n/app/example2.com-0.pem example2.com"

	got, err := h.generateCrtListContent(certificates)
	assert.NoError(t, err)
	assert.Equalf(t, want, string(got), "generateCrtListContent(%v)", certificates)
}

func Test_haproxy_generateCrtListContent_SuccessWithSpecificDomain(t *testing.T) {
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			Path:               "/app",
			AddPem:             true,
			OnlyMatchedDomains: true,
			SpecificDomains: []ConfigSpecificDomain{
				{Identifier: "example.com", Domains: types.Domains{"example.com"}},
			},
		},
	}
	fsStorage := &fs{cfg: cfg.ConfigFs}
	h := haproxy{
		fsStorage: fsStorage,
		cfg:       cfg,
	}
	certificates := types.Certificates{
		{Identifier: "example.com-0", Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: "example2.com-0", Domains: types.Domains{"example2.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	want := "\n/app/example.com.pem example.com"

	got, err := h.generateCrtListContent(certificates)
	assert.NoError(t, err)
	assert.Equalf(t, want, string(got), "generateCrtListContent(%v)", certificates)
}

func Test_haproxy_generateCrtListContent_FailedParse(t *testing.T) {
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			Path:   "/app",
			AddPem: true,
		},
	}
	fsStorage := &fs{cfg: cfg.ConfigFs}
	h := haproxy{
		fsStorage: fsStorage,
		cfg:       cfg,
	}
	certificates := types.Certificates{
		{Identifier: "example.com-0", Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	haproxyCrtListTmplOld := haproxyCrtListTmpl
	haproxyCrtListTmpl = `{{- rang $sni, $path := . }}
{{ $path }} {{ $sni }}
{{- end }}`
	got, err := h.generateCrtListContent(certificates)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template: haproxy_crt_list:1: function \"rang\" not defined")
	assert.Nil(t, got)
	haproxyCrtListTmpl = haproxyCrtListTmplOld
}

func Test_haproxy_generateCrtListContent_FailedExecute(t *testing.T) {
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			Path:   "/app",
			AddPem: true,
		},
	}
	fsStorage := &fs{cfg: cfg.ConfigFs}
	h := haproxy{
		fsStorage: fsStorage,
		cfg:       cfg,
	}
	certificates := types.Certificates{
		{Identifier: "example.com-0", Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	haproxyCrtListTmplOld := haproxyCrtListTmpl
	haproxyCrtListTmpl = `{{- range $sni, $path := . }}
{{ $path.Test }} {{ $sni }}
{{- end }}`
	got, err := h.generateCrtListContent(certificates)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template: haproxy_crt_list:2:8: executing")
	assert.Nil(t, got)
	haproxyCrtListTmpl = haproxyCrtListTmplOld
}

func Test_haproxy_generateCrtListFile_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	certificates := types.Certificates{
		{Identifier: "example.com", Key: []byte("key"), Certificate: []byte("certificate")},
	}
	file, _ := ctx.Fs.Create("/app/test.txt")

	gomock.InOrder(
		// crt list
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, errors.New("fail")),
		fsMock.EXPECT().Mkdir(gomock.Any(), gomock.Any()).Times(1).Return(nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),

		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(file, nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),
	)
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			AddPem: true,
			Path:   "/app",
		},
	}
	checksum := appFs.NewChecksum(fsMock)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: fsMock, checksum: checksum}
	storage := &haproxy{
		fs:        fsMock,
		cfg:       cfg,
		fsStorage: fsStorage,

		checksum: checksum,
	}
	isChanged, err := storage.generateCrtListFile(certificates)
	assert.NoError(t, err)
	assert.True(t, isChanged)
}

func Test_haproxy_generateCrtListFile_FailedCreateDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	certificates := types.Certificates{
		{Identifier: "example.com", Key: []byte("key"), Certificate: []byte("certificate")},
	}

	gomock.InOrder(
		// crt list
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, errors.New("fail")),
		fsMock.EXPECT().Mkdir(gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail")),
	)
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			AddPem: true,
			Path:   "/app",
		},
	}
	checksum := appFs.NewChecksum(fsMock)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: fsMock, checksum: checksum}
	storage := &haproxy{
		fs:        fsMock,
		cfg:       cfg,
		fsStorage: fsStorage,

		checksum: checksum,
	}
	isChanged, err := storage.generateCrtListFile(certificates)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create dir /app")
	assert.False(t, isChanged)
}

func Test_haproxy_generateCrtListFile_FailedGenerateCrtListContent(t *testing.T) {
	ctx := context.TestContext(nil)
	certificates := types.Certificates{
		{Identifier: "example.com", Key: []byte("key"), Certificate: []byte("certificate")},
	}
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			AddPem: true,
			Path:   "/app",
		},
	}
	checksum := appFs.NewChecksum(ctx.Fs)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: ctx.Fs, checksum: checksum}
	storage := &haproxy{
		fs:        ctx.Fs,
		cfg:       cfg,
		fsStorage: fsStorage,

		checksum: checksum,
	}
	haproxyCrtListTmplOld := haproxyCrtListTmpl
	haproxyCrtListTmpl = `{{- rang $sni, $path := . }}
{{ $path }} {{ $sni }}
{{- end }}`
	isChanged, err := storage.generateCrtListFile(certificates)
	haproxyCrtListTmpl = haproxyCrtListTmplOld
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate crt file /app")
	assert.False(t, isChanged)
}

func Test_haproxy_generateCrtListFile_FailedWriteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	certificates := types.Certificates{
		{Identifier: "example.com", Key: []byte("key"), Certificate: []byte("certificate")},
	}

	gomock.InOrder(
		// crt list
		fsMock.EXPECT().Stat(gomock.Any()).Times(1).Return(nil, errors.New("fail")),
		fsMock.EXPECT().Mkdir(gomock.Any(), gomock.Any()).Times(1).Return(nil),
		fsMock.EXPECT().Chown(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),

		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("error")),
	)
	cfg := ConfigHaproxy{
		CrtListPath: "/app/crt-list.txt",
		ConfigFs: ConfigFs{
			AddPem: true,
			Path:   "/app",
		},
	}
	checksum := appFs.NewChecksum(fsMock)
	fsStorage := &fs{cfg: cfg.ConfigFs, fs: fsMock, checksum: checksum}
	storage := &haproxy{
		fs:        fsMock,
		cfg:       cfg,
		fsStorage: fsStorage,

		checksum: checksum,
	}
	isChanged, err := storage.generateCrtListFile(certificates)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write crt file /app/crt-list.tx")
	assert.False(t, isChanged)
}
