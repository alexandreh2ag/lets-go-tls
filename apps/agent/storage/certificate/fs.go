package certificate

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/hook"
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
	Path                string                     `mapstructure:"path" validate:"required"`
	PrefixFilename      string                     `mapstructure:"prefix_filename"`
	SpecificIdentifiers []ConfigSpecificIdentifier `mapstructure:"specific_identifiers" validate:"unique=Identifier,dive"`
	PostHook            *hook.Hook                 `mapstructure:"post_hook"`
}

type ConfigSpecificIdentifier struct {
	Identifier string        `mapstructure:"identifier" validate:"required"`
	Domains    types.Domains `mapstructure:"domains" validate:"required,min=1"`
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
	return f.GetFilePath(cert.GetKeyFilename())
}

func (f fs) GetCertificatePath(cert *types.Certificate) string {
	return f.GetFilePath(cert.GetCertificateFilename())
}

func (f fs) GetFilePath(filename string) string {
	return filepath.Join(f.cfg.Path, fmt.Sprintf("%s%s", f.cfg.PrefixFilename, filename))
}

func (f fs) Save(certificates types.Certificates, hookChan chan<- *hook.Hook) []error {
	errors := []error{}
	err := f.fs.MkdirAll(f.cfg.Path, 0770)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create dir %s: %v", f.cfg.Path, err))
		return errors
	}

	isChanged := false
	for _, cert := range certificates {
		keyPath := f.GetKeyPath(cert)
		certPath := f.GetCertificatePath(cert)

		if identifier := f.GetSpecificIdentifier(cert); identifier != "" {
			keyPath = f.GetFilePath(types.GetKeyFilename(identifier))
			certPath = f.GetFilePath(types.GetCertificateFilename(identifier))
		}

		if !f.checksum.MustCompareContentWithPath(cert.Key, keyPath) {
			isChanged = true
			err = afero.WriteFile(f.fs, keyPath, cert.Key, 0660)
			if err != nil {
				errors = append(errors, fmt.Errorf("fail to write key %s: %v", keyPath, err))
				continue
			}
		}

		if !f.checksum.MustCompareContentWithPath(cert.Certificate, certPath) {
			isChanged = true
			err = afero.WriteFile(f.fs, certPath, cert.Certificate, 0660)
			if err != nil {
				errors = append(errors, fmt.Errorf("fail to write certificate %s: %v", certPath, err))
				continue
			}
		}

	}

	if isChanged && f.cfg.PostHook != nil {
		hookChan <- f.cfg.PostHook
	}

	return errors
}

func (f fs) Delete(certificates types.Certificates, hookChan chan<- *hook.Hook) []error {
	errors := []error{}

	isChanged := false
	for _, cert := range certificates {
		keyPath := f.GetKeyPath(cert)
		certPath := f.GetCertificatePath(cert)
		if ok, _ := afero.Exists(f.fs, keyPath); ok {
			isChanged = true
			err := f.fs.Remove(keyPath)
			if err != nil {
				errors = append(errors, err)
			}
		}

		if ok, _ := afero.Exists(f.fs, certPath); ok {
			isChanged = true
			err := f.fs.Remove(certPath)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	if isChanged && f.cfg.PostHook != nil {
		hookChan <- f.cfg.PostHook
	}

	return errors
}

func (f fs) GetSpecificIdentifier(cert *types.Certificate) string {
	for _, specificIdentifier := range f.cfg.SpecificIdentifiers {
		if cert.Match(specificIdentifier.Domains) {
			return specificIdentifier.Identifier
		}
	}

	return ""
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
