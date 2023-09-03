package config

type StorageConfig struct {
	Id     string                 `mapstructure:"id" validate:"required,excludesall=!@#$ "`
	Type   string                 `mapstructure:"type" validate:"required,excludesall=!@#$ "`
	Config map[string]interface{} `mapstructure:"config"`
}
