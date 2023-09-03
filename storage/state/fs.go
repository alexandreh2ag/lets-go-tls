package state

import (
	"encoding/json"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/state"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
)

const (
	FsKey = "fs"
)

func init() {
	TypeStorageMapping[FsKey] = createFsStorage
}

var _ state.Storage = &fs{}

type ConfigFs struct {
	Path string `mapstructure:"path" validate:"required"`
}

type fs struct {
	fs       afero.Fs
	checksum *appFs.Checksum
	cfg      ConfigFs
}

func (f fs) Type() string {
	return FsKey
}

func (f fs) Load() (*types.State, error) {
	s := &types.State{Account: &acme.Account{}}
	ok, err := afero.Exists(f.fs, f.cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to check path %s: %v", f.cfg.Path, err)
	}
	if !ok {
		return s, nil
	}
	data, _ := f.fs.Open(f.cfg.Path)
	decoder := json.NewDecoder(data)
	err = decoder.Decode(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json %s: %v", f.cfg.Path, err)
	}
	return s, nil
}

func (f fs) Save(state *types.State) error {
	data, _ := json.Marshal(state)

	if !f.checksum.MustCompareContentWithPath(data, f.cfg.Path) {
		err := afero.WriteFile(f.fs, f.cfg.Path, data, 0660)
		if err != nil {
			return fmt.Errorf("failed to write in %s: %v", f.cfg.Path, err)
		}
	}

	return nil
}

func createFsStorage(ctx context.Context, cfg config.StateConfig) (state.Storage, error) {
	instanceConfig := ConfigFs{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	instance := &fs{fs: ctx.GetFS(), cfg: instanceConfig, checksum: appFs.NewChecksum(ctx.GetFS())}

	return instance, nil
}
