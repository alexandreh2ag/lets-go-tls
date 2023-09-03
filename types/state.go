package types

import (
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
)

type State struct {
	Account      *acme.Account `json:"account,omitempty"`
	Certificates Certificates  `json:"certificates"`
}
