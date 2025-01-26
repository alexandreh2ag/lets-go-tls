package config

type HTTPConfig struct {
	Listen        string    `mapstructure:"listen" validate:"required"`
	MetricsEnable bool      `mapstructure:"metrics_enable"`
	TLS           TLSConfig `mapstructure:"tls"`
}

type TLSConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Listen   string `mapstructure:"listen" validate:"required_if=Enable true"`
	CertPath string `mapstructure:"cert_path" validate:"required_if=Enable true"`
	KeyPath  string `mapstructure:"key_path" validate:"required_if=Enable true"`
}
