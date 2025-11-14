package certificate

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/hook"
	"github.com/alexandreh2ag/lets-go-tls/mapstructure"
	"github.com/alexandreh2ag/lets-go-tls/os"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
)

const (
	HaproxyKey = "haproxy"
)

func init() {
	TypeStorageMapping[HaproxyKey] = createHaproxyStorage
}

var _ certificate.Storage = &haproxy{}

var haproxyCrtListTmpl = `{{- range $sni, $path := . }}
{{ $path }} {{ $sni }}
{{- end }}`

type ConfigHaproxy struct {
	ConfigFs    `mapstructure:",squash"`
	CrtListPath string `mapstructure:"crt_list_path"`
}

type haproxy struct {
	fsStorage *fs
	id        string
	fs        afero.Fs
	cfg       ConfigHaproxy

	uid int
	gid int
}

func (h haproxy) ID() string {
	return h.id
}

func (h haproxy) Save(certificates types.Certificates, hookChan chan<- *hook.Hook) []error {
	certsChanged, errors := h.fsStorage.save(certificates)

	crtListChanged, err := h.generateCrtListFile(certificates)
	if err != nil {
		return append(errors, err)
	}
	if (certsChanged || crtListChanged) && h.cfg.PostHook != nil {
		hookChan <- h.cfg.PostHook
	}
	return errors
}

func (h haproxy) Delete(certificates types.Certificates, hookChan chan<- *hook.Hook) []error {
	_, errors := h.fsStorage.delete(certificates)

	return errors
}

func (h haproxy) generateCrtListContent(certificates types.Certificates) ([]byte, error) {
	data := map[string]string{}
	for _, cert := range certificates {
		pathCfg, skip := h.fsStorage.GetPathConfig(cert)
		if skip {
			continue
		}
		for _, domain := range cert.Domains {
			data[string(domain)] = pathCfg.PemPath
		}
	}

	t, err := template.New("haproxy_crt_list").Parse(haproxyCrtListTmpl)
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	err = t.Execute(&result, data)
	if err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func (h haproxy) generateCrtListFile(certificates types.Certificates) (bool, error) {
	path := filepath.Dir(h.cfg.CrtListPath)
	err := appFs.MkdirAllWithChown(h.fs, path, 0755, h.uid, h.gid)
	if err != nil {
		return false, fmt.Errorf("failed to create dir %s: %v", path, err)
	}
	content, errGenerate := h.generateCrtListContent(certificates)
	if errGenerate != nil {
		return false, fmt.Errorf("failed to generate crt file %s: %v", h.cfg.CrtListPath, errGenerate)
	}
	isChanged, errWrite := h.fsStorage.WriteFile(content, h.cfg.CrtListPath)
	if errWrite != nil {
		return false, fmt.Errorf("failed to write crt file %s: %v", h.cfg.CrtListPath, errWrite)
	}
	return isChanged, nil
}

func createHaproxyStorage(ctx *context.AgentContext, cfg config.StorageConfig) (certificate.Storage, error) {
	instanceConfig := ConfigHaproxy{}
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

	if instanceConfig.CrtListPath == "" {
		instanceConfig.CrtListPath = fmt.Sprintf("%s/crt-list.txt", instanceConfig.Path)
	}

	// Force Pem format for haproxy
	instanceConfig.AddPem = true

	uid := os.GetUserUID(instanceConfig.Owner)
	gid := os.GetGroupUID(instanceConfig.Group)

	instanceFs := &fs{id: cfg.Id, fs: ctx.Fs, cfg: instanceConfig.ConfigFs, checksum: appFs.NewChecksum(ctx.Fs), uid: uid, gid: gid}
	instance := &haproxy{id: cfg.Id, fs: ctx.Fs, cfg: instanceConfig, fsStorage: instanceFs, uid: uid, gid: gid}

	return instance, nil
}
