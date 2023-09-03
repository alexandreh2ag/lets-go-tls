package config

import (
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	got := NewConfig()
	assert.Equal(t, Config{}, got)
}

func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig()
	assert.Equal(t, Config{HTTP: config.HTTPConfig{Listen: "0.0.0.0:8080"}, Interval: time.Minute * 5}, got)
}
