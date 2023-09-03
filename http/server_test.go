package http

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateEcho(t *testing.T) {
	got := CreateEcho()

	assert.NotNil(t, got)
}

func TestGetApiPrefix(t *testing.T) {
	res := GetApiPrefix("certificates")
	assert.Equal(t, "/api/certificates", res)
}
