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
	"path/filepath"
	"testing"
)

func Test_fs_ID(t *testing.T) {
	f := &fs{id: "foo"}
	assert.Equal(t, "foo", f.ID())
}

func Test_createFsStorage(t *testing.T) {
	ctx := context.TestContext(nil)
	storage := &fs{id: "foo", fs: ctx.Fs, cfg: ConfigFs{Path: "/app"}, checksum: appFs.NewChecksum(ctx.Fs)}
	storageSpecificIdentifiers := &fs{
		id: "foo",
		fs: ctx.Fs,
		cfg: ConfigFs{Path: "/app", SpecificIdentifiers: []ConfigSpecificIdentifier{
			{Identifier: "test", Domains: types.Domains{"example.com"}},
			{Identifier: "test2", Domains: types.Domains{"example2.com"}},
		}},
		checksum: appFs.NewChecksum(ctx.Fs),
	}
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
			errContains: "Key: 'ConfigFs.Path' Error:Field validation for 'Path' failed on the 'required' tag",
		},
		{
			name: "SuccessValidateCfgSpecificIdentifiers",
			cfg: config.StorageConfig{
				Id: "foo",
				Config: map[string]interface{}{
					"path": "/app",
					"specific_identifiers": []interface{}{
						map[string]interface{}{
							"identifier": "test",
							"domains":    []string{"example.com"},
						},
						map[string]interface{}{
							"identifier": "test2",
							"domains":    []string{"example2.com"},
						},
					}},
			},
			want:    storageSpecificIdentifiers,
			wantErr: false,
		},
		{
			name: "FailValidateCfgSpecificIdentifiers",
			cfg: config.StorageConfig{
				Id: "foo",
				Config: map[string]interface{}{
					"path": "/app",
					"specific_identifiers": []interface{}{
						map[string]interface{}{
							"identifier": "",
							"domains":    []string{},
						},
					}},
			},
			want:        storage,
			wantErr:     true,
			errContains: "Key: 'ConfigFs.SpecificIdentifiers[0].Identifier' Error:Field validation for 'Identifier' failed on the 'required' tag",
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

func Test_fs_Save_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	identifier1 := "example.com-0"
	identifier2 := "foo.com-0"
	identifierCustom := "foo-custom"
	certificates := types.Certificates{
		{Identifier: identifier1, Domains: types.Domains{"example.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
		{Identifier: identifier2, Domains: types.Domains{"foo.com"}, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	storage := &fs{
		fs: ctx.Fs,
		cfg: ConfigFs{
			Path: "/app",
			SpecificIdentifiers: []ConfigSpecificIdentifier{
				{Identifier: identifierCustom, Domains: types.Domains{"foo.com"}},
			},
		},
		checksum: appFs.NewChecksum(ctx.Fs),
	}
	errs := storage.Save(certificates, make(chan<- *hook.Hook))

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
}

func Test_fs_Save_SuccessWithHook(t *testing.T) {
	ctx := context.TestContext(nil)
	identifier := "example.com"
	certificates := types.Certificates{
		{Identifier: identifier, Key: []byte("key"), Certificate: []byte("certificate")},
	}
	chanHook := make(chan *hook.Hook)
	postHook := &hook.Hook{Cmd: "echo 1"}
	storage := &fs{fs: ctx.Fs, cfg: ConfigFs{Path: "/app", PostHook: postHook}, checksum: appFs.NewChecksum(ctx.Fs)}

	go func() {
		for {
			select {
			case h := <-chanHook:
				assert.Equal(t, postHook, h)
			}
		}
	}()

	errs := storage.Save(certificates, chanHook)
	assert.Len(t, errs, 0)
	contentKey, err := afero.ReadFile(ctx.Fs, filepath.Join(storage.cfg.Path, identifier+".key"))
	assert.NoError(t, err)
	assert.Equal(t, "key", string(contentKey))

	contentCrt, err := afero.ReadFile(ctx.Fs, filepath.Join(storage.cfg.Path, identifier+".crt"))
	assert.NoError(t, err)
	assert.Equal(t, "certificate", string(contentCrt))
}

func Test_fs_Save_FailCreateDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	fsMock.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Times(1).Return(errors.New("error"))
	certificates := types.Certificates{}
	storage := &fs{fs: fsMock, cfg: ConfigFs{Path: "/app"}, checksum: appFs.NewChecksum(fsMock)}
	errs := storage.Save(certificates, make(chan<- *hook.Hook))
	assert.Len(t, errs, 1)
}

func Test_fs_Save_FailWriteKey(t *testing.T) {
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
	storage := &fs{fs: fsMock, cfg: ConfigFs{Path: "/app"}, checksum: appFs.NewChecksum(fsMock)}
	errs := storage.Save(certificates, make(chan<- *hook.Hook))
	assert.Len(t, errs, 1)
}

func Test_fs_Save_FailCertificateKey(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	fsMock := mockAfero.NewMockFs(ctrl)
	certificates := types.Certificates{
		{Identifier: "example.com", Key: []byte("key"), Certificate: []byte("certificate")},
	}
	file, _ := ctx.Fs.Create("/app/test.txt")
	gomock.InOrder(
		fsMock.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Times(1).Return(nil),
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(file, nil),
		fsMock.EXPECT().Open(gomock.Any()).Times(1).Return(nil, errors.New("error")),
		fsMock.EXPECT().OpenFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("error")),
	)
	storage := &fs{fs: fsMock, cfg: ConfigFs{Path: "/app"}, checksum: appFs.NewChecksum(fsMock)}
	errs := storage.Save(certificates, make(chan<- *hook.Hook))
	assert.Len(t, errs, 1)
}

func Test_fs_GetKeyPath(t *testing.T) {
	want := "/app/example.com-0.key"
	cert := &types.Certificate{Identifier: "example.com-0"}
	f := &fs{cfg: ConfigFs{Path: "/app"}}
	assert.Equal(t, want, f.GetKeyPath(cert))
}

func Test_fs_GetKeyPathWithPrefix(t *testing.T) {
	want := "/app/ssl.example.com-0.key"
	cert := &types.Certificate{Identifier: "example.com-0"}
	f := &fs{cfg: ConfigFs{Path: "/app", PrefixFilename: "ssl."}}
	assert.Equal(t, want, f.GetKeyPath(cert))
}

func Test_fs_GetCertificatePath(t *testing.T) {
	want := "/app/example.com-0.crt"
	cert := &types.Certificate{Identifier: "example.com-0"}
	f := &fs{cfg: ConfigFs{Path: "/app"}}
	assert.Equal(t, want, f.GetCertificatePath(cert))
}

func Test_fs_GetCertificatePathWithPrefix(t *testing.T) {
	want := "/app/ssl.example.com-0.crt"
	cert := &types.Certificate{Identifier: "example.com-0"}
	f := &fs{cfg: ConfigFs{Path: "/app", PrefixFilename: "ssl."}}
	assert.Equal(t, want, f.GetCertificatePath(cert))
}

func Test_fs_GetFilePath(t *testing.T) {
	want := "/app/example.com-0.crt"
	filename := "example.com-0.crt"
	f := &fs{cfg: ConfigFs{Path: "/app"}}
	assert.Equal(t, want, f.GetFilePath(filename))
}

func Test_fs_GetFilePathWithPrefix(t *testing.T) {
	want := "/app/ssl.example.com-0.crt"
	filename := "example.com-0.crt"
	f := &fs{cfg: ConfigFs{Path: "/app", PrefixFilename: "ssl."}}
	assert.Equal(t, want, f.GetFilePath(filename))
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
	f := &fs{fs: fsMock, cfg: ConfigFs{Path: "/app", PostHook: postHook}}
	assert.Equalf(t, want, f.Delete(certificates, chanHook), "Delete(%v)", certificates)
}

func Test_fs_GetSpecificIdentifier(t *testing.T) {

	tests := []struct {
		name string
		cfg  ConfigFs
		cert *types.Certificate
		want string
	}{
		{
			name: "GetSpecificIdentifier",
			cert: &types.Certificate{Identifier: "example.com-0", Domains: types.Domains{"example.com"}},
			cfg:  ConfigFs{SpecificIdentifiers: []ConfigSpecificIdentifier{{Identifier: "foo", Domains: types.Domains{"example.com"}}}},
			want: "foo",
		},
		{
			name: "empty",
			cert: &types.Certificate{Identifier: "example.com-0", Domains: types.Domains{"example.com"}},
			cfg:  ConfigFs{SpecificIdentifiers: []ConfigSpecificIdentifier{{Identifier: "foo", Domains: types.Domains{"foo.com"}}}},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fs{cfg: tt.cfg}
			assert.Equalf(t, tt.want, f.GetSpecificIdentifier(tt.cert), "GetSpecificIdentifier(%v)", tt.cert)
		})
	}
}
