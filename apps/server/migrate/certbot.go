package migrate

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-jose/go-jose/v4"
	"github.com/spf13/afero"
)

func MigrateCertbot(ctx *context.ServerContext, certbotdir string) (*types.State, error) {
	certbotLiveDir := path.Join(certbotdir, "live")
	certbotAccountDir := path.Join(certbotdir, "accounts")

	certificates, errReadCert := ReadCertbotCertificates(ctx, certbotLiveDir)
	if errReadCert != nil {
		return nil, errReadCert
	}

	account, errReadAccount := ReadCertbotAccount(ctx, certbotAccountDir)
	if errReadAccount != nil {
		return nil, errReadAccount
	}
	return &types.State{Account: account, Certificates: certificates}, nil
}

func ReadCertbotCertificates(ctx *context.ServerContext, dirPath string) (types.Certificates, error) {
	certificates := types.Certificates{}
	if ok, _ := afero.Exists(ctx.Fs, dirPath); !ok {
		return certificates, fmt.Errorf("directory %s does not exist", dirPath)
	}
	err := afero.Walk(ctx.Fs, dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || path == dirPath {
			return nil
		}
		cert := &types.Certificate{
			Identifier: fmt.Sprintf("%s-%v", filepath.Base(path), 0),
			Main:       filepath.Base(path),
		}

		certRaw, errReadCert := afero.ReadFile(ctx.Fs, filepath.Join(path, "fullchain.pem"))
		if errReadCert != nil {
			return errReadCert
		}

		keyRaw, errReadKey := afero.ReadFile(ctx.Fs, filepath.Join(path, "privkey.pem"))
		if errReadKey != nil {
			return errReadKey
		}

		x509Cert, errParse := types.GetX509Certificate(certRaw)
		if errParse != nil {
			return fmt.Errorf("failed to decode certificate pem %s: %v", filepath.Join(path, "fullchain.pem"), errParse)
		}
		domains := types.Domains{}
		for _, domain := range x509Cert.DNSNames {
			domains = append(domains, types.Domain(domain))
		}
		cert.Certificate = certRaw
		cert.Key = keyRaw
		cert.Domains = domains
		cert.ExpirationDate = x509Cert.NotAfter
		cert.ObtainFailCount = 0
		cert.ObtainFailDate = time.Time{}
		certificates = append(certificates, cert)

		return nil
	})
	if err != nil {
		return certificates, err
	}
	return certificates, nil
}

func ReadCertbotAccount(ctx *context.ServerContext, dirPath string) (*acme.Account, error) {
	accountPath := ""

	if ok, _ := afero.Exists(ctx.Fs, dirPath); !ok {
		return nil, fmt.Errorf("directory %s does not exist", dirPath)
	}

	_ = afero.Walk(ctx.Fs, dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Base(path) == "regr.json" {
			accountPath = filepath.Dir(path)
		}

		return nil
	})

	if accountPath == "" {
		return nil, errors.New("no account not found")
	}
	ctx.Logger.Info(fmt.Sprintf("Reading account file path %s", accountPath))

	// regr.json exist because found in walk
	regrRaw, _ := afero.ReadFile(ctx.Fs, path.Join(accountPath, "regr.json"))

	keyRaw, errReadkey := afero.ReadFile(ctx.Fs, path.Join(accountPath, "private_key.json"))
	if errReadkey != nil {
		return nil, errReadkey
	}

	reg := &registration.Resource{}
	errParseReg := json.Unmarshal(regrRaw, reg)
	if errParseReg != nil {
		return nil, fmt.Errorf("failed to unmarshal acme regr.json file: %w", errParseReg)
	}

	var jwk jose.JSONWebKey
	errParseKey := json.Unmarshal(keyRaw, &jwk)
	if errParseKey != nil {
		return nil, fmt.Errorf("failed to unmarshal acme private_key.json file: %w", errParseKey)
	}
	key, ok := jwk.Key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("failed to parse private key with RSA type")
	}

	return &acme.Account{Email: "", Registration: reg, Key: x509.MarshalPKCS1PrivateKey(key)}, nil
}
