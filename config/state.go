package config

type StateConfig struct {
	Type   string                 `mapstructure:"type" validate:"required,excludesall=!@#$ "`
	Config map[string]interface{} `mapstructure:"config"`
}
