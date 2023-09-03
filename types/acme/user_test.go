package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"github.com/go-acme/lego/v4/registration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAccount(t *testing.T) {
	email := "dev@example.com"

	got, err := NewAccount(email)
	assert.NoError(t, err)
	assert.Equal(t, email, got.Email)
	assert.NotEmpty(t, got.Key)
}

func TestAccount_GetEmail(t *testing.T) {
	email := "dev@example.com"

	account := &Account{Email: email}
	assert.Equal(t, email, account.GetEmail())
}

func TestAccount_GetRegistration(t *testing.T) {
	reg := &registration.Resource{}
	account := &Account{Registration: reg}
	assert.Equal(t, reg, account.GetRegistration())
}

func TestAccount_GetPrivateKey_Success(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	assert.NoError(t, err)
	key := x509.MarshalPKCS1PrivateKey(privateKey)
	account := &Account{Key: key}
	assert.Equal(t, privateKey, account.GetPrivateKey())
}

func TestAccount_GetPrivateKey_Fail(t *testing.T) {
	account := &Account{}
	defer func() {
		if r := recover(); r != nil {
			assert.True(t, true)
		} else {
			t.Errorf("GetPrivateKey should have panicked")
		}
	}()
	_ = account.GetPrivateKey()
}
