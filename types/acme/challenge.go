package acme

import "github.com/go-acme/lego/v4/challenge"

const (
	TypeDNS01  = "dns-01"
	TypeHTTP01 = "http-01"
)

type Challenges = map[string]Challenge

type Challenge interface {
	challenge.Provider
	ID() string
	Type() string
}
