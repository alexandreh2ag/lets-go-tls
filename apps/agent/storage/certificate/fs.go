package certificate

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/hook"
	"github.com/alexandreh2ag/lets-go-tls/os"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"path/filepath"
	"slices"
)

const (
	FsKey = "fs"
)

func init() {
	TypeStorageMapping[FsKey] = createFsStorage
}

var _ certificate.Storage = &fs{}

type ConfigFs struct {
	Path               string                 `mapstructure:"path" validate:"required"`
	PrefixFilename     string                 `mapstructure:"prefix_filename"`
	OnlyMatchedDomains bool                   `mapstructure:"only_matched_domains"`
	SpecificDomains    []ConfigSpecificDomain `mapstructure:"specific_domains" validate:"duplicate_path,dive"`
	PostHook           *hook.Hook             `mapstructure:"post_hook"`

	AddPem bool `mapstructure:"add_pem"`

	Owner string `mapstructure:"owner"`
	Group string `mapstructure:"group"`
}

type ConfigSpecificDomain struct {
	Identifier string        `mapstructure:"identifier" validate:"required"`
	Path       string        `mapstructure:"path"`
	Domains    types.Domains `mapstructure:"domains" validate:"required,min=1"`
}

type fs struct {
	id       string
	fs       afero.Fs
	checksum *appFs.Checksum
	cfg      ConfigFs

	uid int
	gid int
}

type certificatePathCfg struct {
	CertPath string
	KeyPath  string
	PemPath  string
}

func (f fs) ID() string {
	return f.id
}

func (f fs) GetKeyPath(cert *types.Certificate) string {
	return f.GetFilePath(f.cfg.Path, cert.GetKeyFilename())
}

func (f fs) GetCertificatePath(cert *types.Certificate) string {
	return f.GetFilePath(f.cfg.Path, cert.GetCertificateFilename())
}

func (f fs) GetPemPath(cert *types.Certificate) string {
	return f.GetFilePath(f.cfg.Path, cert.GetPemFilename())
}

func (f fs) GetFilePath(path, filename string) string {
	if path == "" {
		path = f.cfg.Path
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(f.cfg.Path, path)
	}

	return filepath.Join(path, fmt.Sprintf("%s%s", f.cfg.PrefixFilename, filename))
}

func (f fs) GetPathConfig(cert *types.Certificate) (certificatePathCfg, bool) {
	pathCfg := certificatePathCfg{
		KeyPath:  f.GetKeyPath(cert),
		CertPath: f.GetCertificatePath(cert),
		PemPath:  f.GetPemPath(cert),
	}

	if specificDomainCfg := f.GetSpecificDomainConfig(cert); specificDomainCfg != nil {
		pathCfg.KeyPath = f.GetFilePath(specificDomainCfg.Path, types.GetKeyFilename(specificDomainCfg.Identifier))
		pathCfg.CertPath = f.GetFilePath(specificDomainCfg.Path, types.GetCertificateFilename(specificDomainCfg.Identifier))
		pathCfg.PemPath = f.GetFilePath(specificDomainCfg.Path, types.GetPemFilename(specificDomainCfg.Identifier))
	} else if f.cfg.OnlyMatchedDomains && len(f.cfg.SpecificDomains) > 0 {
		return certificatePathCfg{}, true
	}
	return pathCfg, false
}

func (f fs) Save(certificates types.Certificates, hookChan chan<- *hook.Hook) []error {
	isChanged, errors := f.save(certificates)

	if isChanged && f.cfg.PostHook != nil {
		hookChan <- f.cfg.PostHook
	}

	return errors
}

func (f fs) save(certificates types.Certificates) (bool, []error) {
	errors := []error{}

	isChanged := false
	for _, cert := range certificates {

		pathCfg, skip := f.GetPathConfig(cert)
		if skip {
			continue
		}

		path := filepath.Dir(pathCfg.KeyPath)
		err := appFs.MkdirAllWithChown(f.fs, path, 0755, f.uid, f.gid)
		if err != nil {
			errors = append(errors, fmt.Errorf("unable to create dir %s: %v", path, err))
			return false, errors
		}

		keyChanged, errWriteKey := f.WriteFile(cert.Key, pathCfg.KeyPath)
		if errWriteKey != nil {
			errors = append(errors, errWriteKey)
			continue
		}

		certChanged, errWriteCert := f.WriteFile(cert.Certificate, pathCfg.CertPath)
		if errWriteCert != nil {
			errors = append(errors, errWriteCert)
			continue
		}

		pemChanged := false
		if f.cfg.AddPem {
			var errWritePem error
			pemChanged, errWritePem = f.WriteFile(cert.GetPemContent(), pathCfg.PemPath)
			if errWritePem != nil {
				errors = append(errors, errWritePem)
				continue
			}
		}

		if keyChanged || certChanged || pemChanged {
			isChanged = true
		}

	}

	return isChanged, errors
}

func (f fs) WriteFile(content []byte, path string) (bool, error) {
	if !f.checksum.MustCompareContentWithPath(content, path) {
		err := afero.WriteFile(f.fs, path, content, 0660)
		if err != nil {
			return false, fmt.Errorf("fail to write file %s: %v", path, err)
		}

		err = f.fs.Chown(path, f.uid, f.gid)
		if err != nil {
			return false, fmt.Errorf("fail to chown %s: %v", path, err)
		}
		return true, nil
	}
	return false, nil
}

func (f fs) Delete(certificates types.Certificates, hookChan chan<- *hook.Hook) []error {
	isChanged, errors := f.delete(certificates)

	if isChanged && f.cfg.PostHook != nil {
		hookChan <- f.cfg.PostHook
	}

	return errors
}

func (f fs) delete(certificates types.Certificates) (bool, []error) {
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

	return isChanged, errors
}

func (f fs) GetSpecificDomainConfig(cert *types.Certificate) *ConfigSpecificDomain {
	for _, specificDomainCfg := range f.cfg.SpecificDomains {
		if cert.Match(specificDomainCfg.Domains) {
			return &specificDomainCfg
		}
	}

	return nil
}

func createFsStorage(ctx *context.AgentContext, cfg config.StorageConfig) (certificate.Storage, error) {
	instanceConfig := ConfigFs{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	_ = validate.RegisterValidation("duplicate_path", ValidateSpecificDomainsCfg())
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	uid := os.GetUserUID(instanceConfig.Owner)
	gid := os.GetGroupUID(instanceConfig.Group)

	instance := &fs{id: cfg.Id, fs: ctx.Fs, cfg: instanceConfig, checksum: appFs.NewChecksum(ctx.Fs), uid: uid, gid: gid}

	return instance, nil
}

func ValidateSpecificDomainsCfg() func(level validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		specificDomainsCfg := fl.Field().Interface().([]ConfigSpecificDomain)
		uniquesPath := []string{}
		for _, domainCfg := range specificDomainsCfg {
			path := filepath.Join(domainCfg.Path, domainCfg.Identifier)
			if slices.Contains(uniquesPath, path) {
				return false
			}
			uniquesPath = append(uniquesPath, path)
		}

		return true
	}
}
