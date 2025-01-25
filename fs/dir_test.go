package fs

import (
	"errors"
	mockAfero "github.com/alexandreh2ag/lets-go-tls/mocks/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"os"
	"path/filepath"
	"testing"
)

func TestMkdirAllWithChown(t *testing.T) {

	tests := []struct {
		name    string
		path    string
		mockFn  func(fs *mockAfero.MockFs)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "successAllDirExists",
			path: "/tmp/test",
			mockFn: func(fs *mockAfero.MockFs) {
				fs.EXPECT().Stat(gomock.Eq("/tmp")).Times(1).Return(nil, nil)
				fs.EXPECT().Stat(gomock.Eq("/tmp/test")).Times(1).Return(nil, nil)
			},
			wantErr: assert.NoError,
		},
		{
			name: "successAllDirExistsWithRelativePath",
			path: "test",
			mockFn: func(fs *mockAfero.MockFs) {
				dir, _ := os.MkdirTemp("", "")
				_ = os.Chdir(dir)
				fs.EXPECT().Stat(gomock.Eq("/tmp")).Times(1).Return(nil, nil)
				fs.EXPECT().Stat(gomock.Eq(dir)).Times(1).Return(nil, nil)
				fs.EXPECT().Stat(gomock.Eq(filepath.Join(dir, "test"))).Times(1).Return(nil, nil)
			},
			wantErr: assert.NoError,
		},
		{
			name: "FailCheckAbsPath",
			path: "test",
			mockFn: func(fs *mockAfero.MockFs) {
				dir, _ := os.MkdirTemp("", "")
				_ = os.Chdir(dir)
				_ = os.RemoveAll(dir)
			},
			wantErr: assert.Error,
		},
		{
			name: "successLastDirNotExists",
			path: "/tmp/cert/test",
			mockFn: func(fs *mockAfero.MockFs) {
				gomock.InOrder(
					fs.EXPECT().Stat(gomock.Eq("/tmp")).Times(1).Return(nil, nil),
					fs.EXPECT().Stat(gomock.Eq("/tmp/cert")).Times(1).Return(nil, nil),
					fs.EXPECT().Stat(gomock.Eq("/tmp/cert/test")).Times(1).Return(nil, errors.New("not found")),
					fs.EXPECT().Mkdir(gomock.Eq("/tmp/cert/test"), gomock.Eq(os.FileMode(0770))).Times(1).Return(nil),
					fs.EXPECT().Chown(gomock.Eq("/tmp/cert/test"), gomock.Eq(0), gomock.Eq(0)).Times(1).Return(nil),
				)
			},
			wantErr: assert.NoError,
		},
		{
			name: "FailMkdir",
			path: "/tmp/test",
			mockFn: func(fs *mockAfero.MockFs) {
				gomock.InOrder(
					fs.EXPECT().Stat(gomock.Eq("/tmp")).Times(1).Return(nil, nil),
					fs.EXPECT().Stat(gomock.Eq("/tmp/test")).Times(1).Return(nil, errors.New("not found")),
					fs.EXPECT().Mkdir(gomock.Eq("/tmp/test"), gomock.Eq(os.FileMode(0770))).Times(1).Return(errors.New("fail")),
				)
			},
			wantErr: assert.Error,
		},
		{
			name: "FailChown",
			path: "/tmp/test",
			mockFn: func(fs *mockAfero.MockFs) {
				gomock.InOrder(
					fs.EXPECT().Stat(gomock.Eq("/tmp")).Times(1).Return(nil, nil),
					fs.EXPECT().Stat(gomock.Eq("/tmp/test")).Times(1).Return(nil, errors.New("not found")),
					fs.EXPECT().Mkdir(gomock.Eq("/tmp/test"), gomock.Eq(os.FileMode(0770))).Times(1).Return(nil),
					fs.EXPECT().Chown(gomock.Eq("/tmp/test"), gomock.Eq(0), gomock.Eq(0)).Times(1).Return(errors.New("fail")),
				)
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fsMock := mockAfero.NewMockFs(ctrl)
			tt.mockFn(fsMock)
			tt.wantErr(t, MkdirAllWithChown(fsMock, tt.path, 0770, 0, 0))
		})
	}
}
