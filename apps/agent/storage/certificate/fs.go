package certificate

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"path/filepath"
)

const (
	FsKey = "fs"
)

func init() {
	TypeStorageMapping[FsKey] = createFsStorage
}

var _ certificate.Storage = &fs{}

type ConfigFs struct {
	Path           string `mapstructure:"path" validate:"required"`
	PrefixFilename string `mapstructure:"prefix_filename"`
}

type fs struct {
	id       string
	fs       afero.Fs
	checksum *appFs.Checksum
	cfg      ConfigFs
}

func (f fs) ID() string {
	return f.id
}

func (f fs) GetKeyPath(cert *types.Certificate) string {
	return filepath.Join(f.cfg.Path, fmt.Sprintf("%s%s", f.cfg.PrefixFilename, cert.GetKeyFilename()))
}

func (f fs) GetCertificatePath(cert *types.Certificate) string {
	return filepath.Join(f.cfg.Path, fmt.Sprintf("%s%s", f.cfg.PrefixFilename, cert.GetCertificateFilename()))
}

func (f fs) Save(certificates types.Certificates) []error {
	errors := []error{}
	err := f.fs.MkdirAll(f.cfg.Path, 0770)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create dir %s: %v", f.cfg.Path, err))
		return errors
	}

	for _, cert := range certificates {
		keyPath := f.GetKeyPath(cert)
		certPath := f.GetCertificatePath(cert)
		if !f.checksum.MustCompareContentWithPath(cert.Key, keyPath) {
			err = afero.WriteFile(f.fs, keyPath, cert.Key, 0660)
			if err != nil {
				errors = append(errors, fmt.Errorf("fail to write key %s: %v", keyPath, err))
				continue
			}
		}

		if !f.checksum.MustCompareContentWithPath(cert.Certificate, certPath) {
			err = afero.WriteFile(f.fs, certPath, cert.Certificate, 0660)
			if err != nil {
				errors = append(errors, fmt.Errorf("fail to write certificate %s: %v", certPath, err))
				continue
			}
		}

	}

	return errors
}

func (f fs) Delete(certificates types.Certificates) []error {
	errors := []error{}
	for _, cert := range certificates {
		keyPath := f.GetKeyPath(cert)
		certPath := f.GetCertificatePath(cert)
		if ok, _ := afero.Exists(f.fs, keyPath); ok {
			err := f.fs.Remove(keyPath)
			if err != nil {
				errors = append(errors, err)
			}
		}

		if ok, _ := afero.Exists(f.fs, certPath); ok {
			err := f.fs.Remove(certPath)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	return errors
}

func createFsStorage(ctx *context.AgentContext, cfg config.StorageConfig) (certificate.Storage, error) {
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

	instance := &fs{id: cfg.Id, fs: ctx.Fs, cfg: instanceConfig, checksum: appFs.NewChecksum(ctx.Fs)}

	return instance, nil
}
