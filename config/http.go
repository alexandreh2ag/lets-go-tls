package config

type HTTPConfig struct {
	Listen        string `mapstructure:"listen" validate:"required"`
	MetricsEnable bool   `mapstructure:"metrics_enable"`
}
