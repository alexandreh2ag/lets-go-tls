package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/go-acme/lego/v4/registration"
)

var _ registration.User = &Account{}

type Account struct {
	Email        string
	Registration *registration.Resource
	Key          []byte
}

func (a *Account) GetEmail() string {
	return a.Email
}
func (a *Account) GetRegistration() *registration.Resource {
	return a.Registration
}
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	privateKey, err := x509.ParsePKCS1PrivateKey(a.Key)
	if err != nil {
		panic(fmt.Errorf("failed to parse private key: %v", err))
	}
	return privateKey
}

func NewAccount(email string) (*Account, error) {
	account := &Account{Email: email}
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}
	account.Key = x509.MarshalPKCS1PrivateKey(privateKey)
	return account, nil
}
