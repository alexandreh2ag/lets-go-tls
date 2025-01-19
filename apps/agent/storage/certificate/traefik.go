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
	"gopkg.in/yaml.v2"
	"path/filepath"
)

const (
	TraefikV2Key = "traefikV2"
	TraefikV3Key = "traefikV3"
)

func init() {
	TypeStorageMapping[TraefikV2Key] = createTraefikStorage
	TypeStorageMapping[TraefikV3Key] = createTraefikStorage
}

var _ certificate.Storage = &traefik{}

type ConfigTraefik struct {
	Path           string `mapstructure:"path" validate:"required"`
	PrefixFilename string `mapstructure:"prefix_filename"`

	Owner string `mapstructure:"owner"`
	Group string `mapstructure:"group"`
}

type traefik struct {
	id       string
	fs       afero.Fs
	checksum *appFs.Checksum
	cfg      ConfigTraefik

	uid int
	gid int
}

func (t traefik) ID() string {
	return t.id
}

func (t traefik) GetFilePath(cert *types.Certificate) string {
	return filepath.Join(t.cfg.Path, fmt.Sprintf("%s%s.%s", t.cfg.PrefixFilename, cert.Identifier, "yml"))
}

func (t traefik) Save(certificates types.Certificates, _ chan<- *hook.Hook) []error {
	errors := []error{}
	err := t.fs.MkdirAll(t.cfg.Path, 0770)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create dir %s: %v", t.cfg.Path, err))
		return errors
	}
	for _, cert := range certificates {
		err = t.WriteCertFile(cert)
		if err != nil {
			errors = append(errors, err)
			continue
		}
	}

	return errors
}

func (t traefik) Delete(certificates types.Certificates, _ chan<- *hook.Hook) []error {
	errors := []error{}
	for _, cert := range certificates {

		err := t.fs.Remove(t.GetFilePath(cert))
		if err != nil {
			errors = append(errors, err)
			continue
		}
	}
	return errors
}

func (t traefik) WriteCertFile(cert *types.Certificate) error {
	path := t.GetFilePath(cert)
	data := map[string]map[string][]map[string]string{}

	certData := map[string]string{
		"keyFile":  string(cert.Key),
		"certFile": string(cert.Certificate),
	}

	data["tls"] = map[string][]map[string]string{"certificates": {certData}}
	content, _ := yaml.Marshal(data)

	if !t.checksum.MustCompareContentWithPath(content, path) {
		errWrite := afero.WriteFile(t.fs, path, content, 0660)
		if errWrite != nil {
			return fmt.Errorf("fail to write %s: %v", path, errWrite)
		}
		errChown := t.fs.Chown(path, t.uid, t.gid)
		if errChown != nil {
			return fmt.Errorf("fail to chown %s: %v", path, errChown)
		}

	}
	return nil
}

func createTraefikStorage(ctx *context.AgentContext, cfg config.StorageConfig) (certificate.Storage, error) {
	instanceConfig := ConfigTraefik{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	uid := os.GetUserUID(instanceConfig.Owner)
	gid := os.GetGroupUID(instanceConfig.Group)

	instance := &traefik{id: cfg.Id, fs: ctx.Fs, cfg: instanceConfig, checksum: appFs.NewChecksum(ctx.Fs), uid: uid, gid: gid}

	return instance, nil
}
