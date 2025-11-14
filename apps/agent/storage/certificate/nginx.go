package certificate

import (
	"log/slog"
	"path/filepath"

	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appFs "github.com/alexandreh2ag/lets-go-tls/fs"
	"github.com/alexandreh2ag/lets-go-tls/hook"
	"github.com/alexandreh2ag/lets-go-tls/mapstructure"
	nginxParser "github.com/alexandreh2ag/lets-go-tls/nginx"
	"github.com/alexandreh2ag/lets-go-tls/os"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/go-playground/validator/v10"
)

const (
	NginxKey = "nginx"
)

func init() {
	TypeStorageMapping[NginxKey] = createNginxStorage
}

var _ certificate.Storage = &nginx{}

type ConfigNginx struct {
	ConfigFs     `mapstructure:",squash"`
	NginxCfgPath string `mapstructure:"nginx_cfg_path" validate:"required"`
}

type nginx struct {
	fsStorage *fs
	id        string
	cfg       ConfigNginx
	logger    *slog.Logger

	uid int
	gid int
}

func (n nginx) ID() string {
	return n.id
}

func (n nginx) Save(certificates types.Certificates, hookChan chan<- *hook.Hook) []error {
	errors := []error{}
	vhostConfigs, err := nginxParser.ParseConfig(n.logger, n.cfg.NginxCfgPath)
	if err != nil {
		return []error{err}
	}

	certsChanged := false
	for _, vhostConfig := range vhostConfigs {
		cert := certificates.Match(vhostConfig.ServerName, true)
		if cert != nil {
			keyChanged, errWriteKey := n.fsStorage.WriteFile(cert.Key, vhostConfig.KeyPath)
			if errWriteKey != nil {
				errors = append(errors, errWriteKey)
				continue
			}
			certChanged, errWriteCert := n.fsStorage.WriteFile(cert.Certificate, vhostConfig.CertPath)
			if errWriteCert != nil {
				errors = append(errors, errWriteCert)
				continue
			}

			if keyChanged || certChanged {
				certsChanged = true
			}
			continue
		}
	}

	if certsChanged && n.cfg.PostHook != nil {
		hookChan <- n.cfg.PostHook
	}
	return errors
}

func (n nginx) Delete(_ types.Certificates, _ chan<- *hook.Hook) []error {
	return []error{}
}

func createNginxStorage(ctx *context.AgentContext, cfg config.StorageConfig) (certificate.Storage, error) {
	instanceConfig := ConfigNginx{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	instanceConfig.ConfigFs.Path = filepath.Dir(instanceConfig.NginxCfgPath)
	validate := validator.New()
	_ = validate.RegisterValidation("duplicate_path", ValidateSpecificDomainsCfg())
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	uid := os.GetUserUID(instanceConfig.Owner)
	gid := os.GetGroupUID(instanceConfig.Group)

	checksum := appFs.NewChecksum(ctx.Fs)
	instanceFs := &fs{id: cfg.Id, fs: ctx.Fs, cfg: instanceConfig.ConfigFs, checksum: checksum, uid: uid, gid: gid}
	instance := &nginx{id: cfg.Id, logger: ctx.Logger, cfg: instanceConfig, fsStorage: instanceFs, uid: uid, gid: gid}

	return instance, nil
}
