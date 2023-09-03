package cache

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateCache_Success(t *testing.T) {
	cfg := config.CacheConfig{Type: memoryKey, Config: map[string]interface{}{}}
	got, err := CreateCache(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateCache_Fail(t *testing.T) {
	cfg := config.CacheConfig{Type: "wrong", Config: map[string]interface{}{}}
	got, err := CreateCache(cfg)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cache type 'wrong' does not exist")
}
