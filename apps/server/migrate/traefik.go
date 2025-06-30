package migrate

import (
	"encoding/json"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/registration"
	"github.com/spf13/afero"
	"time"
)

func MigrateTraefik(ctx *context.ServerContext, acmePath string) (*types.State, error) {

	certificates, errReadCert := ReadTraefikData(ctx, acmePath)
	if errReadCert != nil {
		return nil, errReadCert
	}
	return certificates, nil
}

func ReadTraefikData(ctx *context.ServerContext, acmePath string) (*types.State, error) {
	// Define a struct to parse Traefik ACME JSON data
	type TraefikACMEData struct {
		Account struct {
			Email        string                 `json:"Email"`
			Registration *registration.Resource `json:"Registration"`
			PrivateKey   []byte                 `json:"PrivateKey"`
		} `json:"Account"`
		Certificates []struct {
			Domain struct {
				Main string        `json:"main"`
				SANs types.Domains `json:"sans"`
			} `json:"domain"`
			Certificate []byte `json:"certificate"`
			Key         []byte `json:"key"`
		} `json:"Certificates"`
	}

	// Initialize the struct
	var acmeData map[string]TraefikACMEData

	fileContent, err := afero.ReadFile(ctx.Fs, acmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Traefik ACME file: %w", err)
	}

	// Parse JSON into the struct
	err = json.Unmarshal(fileContent, &acmeData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Traefik ACME file: %w", err)
	}

	account := &acme.Account{}

	// Process each certificate
	certificates := types.Certificates{}
	for _, dataResolver := range acmeData {
		account.Email = dataResolver.Account.Email
		account.Registration = dataResolver.Account.Registration
		account.Key = dataResolver.Account.PrivateKey
		for _, certData := range dataResolver.Certificates {
			x509Cert, errParse := types.GetX509Certificate(certData.Certificate)
			if errParse != nil {
				return nil, errParse
			}
			cert := &types.Certificate{
				Identifier:      fmt.Sprintf("%s-%v", certData.Domain.Main, 0),
				Main:            certData.Domain.Main,
				Domains:         certData.Domain.SANs,
				Certificate:     certData.Certificate,
				Key:             certData.Key,
				ExpirationDate:  x509Cert.NotAfter,
				ObtainFailCount: 0,
				ObtainFailDate:  time.Time{},
			}

			certificates = append(certificates, cert)
		}
	}

	return &types.State{Account: account, Certificates: certificates}, nil
}
