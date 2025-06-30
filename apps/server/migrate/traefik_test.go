package migrate

import (
	"encoding/base64"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	legoAcme "github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/registration"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"time"
)

func TestReadTraefikData(t *testing.T) {
	ctx := context.TestContext(nil)
	acmePath := "/app/letsencrypt/acme.json"

	accountKeyBase64 := "MIIEowIBAAKCAQEA9pbYlW2rfzxM9BnhKnWXWQvOaSofpbgpLnDZqBb5UJuOexknabPVVjwAJCXyn004Q4J7mjkNsc6xK42Pd8l/2Z1FODh7NJjw/TDBIMn44WQqRF1tZe29WIZBvaW1g7VXkJ4oI93osUf4sqQ+24tUOXYktAVirITk4tIgfiKAV+KI28Cye2CkCmMZtWi/yKoySo+dVLUj1BfZ3BBeDRNVEr2VcNtq6qqZ54phtblcRJPaJ2ZRr77hZJN+DU5Q1IttnJqf1OJkndlRbAdgZ5uHH/k3w+6MlobD1YZPT6bSsSoZMx6YRt14PVWLhuGu2NB/l3JFbGHF2kw3JsQUX5695wIDAQABAoIBAAsHNgqAzCG1na9xFNvkRG+MEqArbRzpPpZAH8vfsyjCvD2PI28bxtfDeTIkrFQpjPGf/U0pbg5csxnCi/8MePZPjU/1XrU8fdKlAo4n2kSiSS2jjAf5+PTjlD8aESMh0he6qwrnmseK83lx1QnWkl67hOq0LvSOQPpP8XoSqvDDL2gwUmP2cG5snd0JHl0matPAkS9J5ios8CmpqBG8TXXh2Q18mLSSlGhUaUt9dSq++Drgugu40EUiM+BpQDdg4YCzJyWsRWZLyFpJ5Ka0x0c5Zgo+kTZkRfTlRZQYSUyXSYxQMnzC6XcMzpIvkOWHW8Vq3APuzSYB773GX9ZP5XkCgYEA/dc0cSGPqeGwXUs1zVuOQc7sJODIjlSZ+Jt8GJ1geQBU8wLCFYkw+Fa6EfJUxk8qhgHpjgelLCv1m8x5fxPeUVafWSug7XqnwSvDpaThS1SNbzOPsyYYSI7RBuVLJz5y949fxziPawf0+Ip6Vv1zKaGKbz+gZFyJyZ8wxRbIaqUCgYEA+K/ZgKYrFHX96ZnjsVUyTQ0HZUAJGvNMaicA8sa/EgiDGtiDLO0yn6btV4uaVMiXy/5GRPikaHO0xD3DYilkGk0WGnswmhy24NyV7H5NDqNyT6BNFDldEe0/8N3WFs1ppY+tzWQgnSbazIuLq5wbR2TcyzAXJM4bXTjIXD0RvJsCgYBf/5BcrJWZaLn7PmNVi6eMKcZS2BPhCAvG4n6XDIWcnUyDLurZnC3B7YKnIN+AtBXf4j9CsI992g4j4rucHJZ6amkPHxnYVge1XWi7b1So2Yw3nudcmC6c2i4hDnxSliKCj11oiXJrxvGVQUGtPmLcPrSNCCIPcpPFhYe7MbeP5QKBgQC+q9VP1hfhsDXmDZ+OL4xJl1u8PcS9bR2BpjYL7mZBJVWgF7E7mQHfdtUy8k56AvKf7W1jJGCp1U0byV8svB+MKt5vn1Q8VRfCgF5S3OOn/Zl/g6CViQXa4NTNkMi84vHFhsvzVnnLBCdt/PAxu5cOvj6ra0vTZFOGviqgzhHzUQKBgC3B1SFn525KuuvCw7KiZOPH7DcHyVI5sbL1FK3/WGFa41Nr+K+kldA2G/lFo8HHHOSmaZzWNbGRp26UjKKG936g4T2+dHCzIC34+w5lcUfYRiEZzYAGAH9n95lHwMr41cCRiChqAUvL7J12w5s2URp7SgUkSBb0WJApFQxKF4vE"
	accountKey, errDecodeAccountKey := base64.StdEncoding.DecodeString(accountKeyBase64)
	assert.NoError(t, errDecodeAccountKey)
	accountRaw := "\"Account\":{\"Email\":\"admin@example.local\",\"Registration\":{\"body\":{\"status\":\"valid\",\"contact\":[\"mailto:admin@example.local\"],\"orders\":\"http://server/acme/name/directory/azerty/orders\"},\"uri\":\"http://server/acme/name/directory/azerty\"},\"PrivateKey\":\"" + accountKeyBase64 + "\",\"KeyType\":\"4096\"}"
	account := &acme.Account{
		Email: "admin@example.local",
		Registration: &registration.Resource{
			Body: legoAcme.Account{Status: "valid", Contact: []string{"mailto:admin@example.local"}, Orders: "http://server/acme/name/directory/azerty/orders"},
			URI:  "http://server/acme/name/directory/azerty",
		},
		Key: accountKey,
	}

	certBase64 := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUI5akNDQVp5Z0F3SUJBZ0lRS0Yxd2piTG9QSEYwMXVXNDU2U2xNakFLQmdncWhrak9QUVFEQWpBOE1SUXcKRWdZRFZRUUtFd3RTWldac1pYUWdSR1YyYkRFa01DSUdBMVVFQXhNYlVtVm1iR1YwSUVSbGRtd2dTVzUwWlhKdApaV1JwWVhSbElFTkJNQjRYRFRJMU1EUXdOakV5TkRBeE5Wb1hEVE0xTURRd056QXdOREV4TlZvd0FEQlpNQk1HCkJ5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCTEdiQTZCa1NTVFBzQkZOakV3L2V4eWJ0MXpuV296NFcxeGoKK2FHQnRIdDFDNGMvZTEwRlM1a0lHdlpvR1phWDFWVmhpUE5uZ29mQVZuSkdwY1AzMWQyamdic3dnYmd3RGdZRApWUjBQQVFIL0JBUURBZ2VBTUIwR0ExVWRKUVFXTUJRR0NDc0dBUVVGQndNQkJnZ3JCZ0VGQlFjREFqQWRCZ05WCkhRNEVGZ1FVM2MyYUJRUUllL0M4Z3Q5MXV4Y2x1SU5HemZRd0h3WURWUjBqQkJnd0ZvQVUzZnR6VGp0UzJiOUUKaW1oU0w3VkY2ZjRBL1hZd0tRWURWUjBSQVFIL0JCOHdIWUlNWkdWMkxtTmxjblF1WVdoaGdnMWtaWFl5TG1ObApjblF1WVdoaE1Cd0dEQ3NHQVFRQmdxUmt4aWhBQVFRTU1Bb0NBUVlFQTJSbGRnUUFNQW9HQ0NxR1NNNDlCQU1DCkEwZ0FNRVVDSUdxRzBRMkR5MitjeE1DZURjck1HUW1ZZldlWC9xbVBFZ3lZRVVzZGoyMGhBaUVBbDIyZnFJQjQKT2J0ekpDdVRFUUtoaWQwMTNSMzhvSUdHb0xDeHJQTzRCVlk9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
	keyBase64 := "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JR0hBZ0VBTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEJHMHdhd0lCQVFRZ21OOVk0THpVTHJTczExNEUKRXF1b2N4R25wNGhZMFRiL0hRbmZJMnhueXllaFJBTkNBQVN4bXdPZ1pFa2t6N0FSVFl4TVAzc2NtN2RjNTFxTQorRnRjWS9taGdiUjdkUXVIUDN0ZEJVdVpDQnIyYUJtV2w5VlZZWWp6WjRLSHdGWnlScVhEOTlYZAotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg=="
	cert, errDecodeCert := base64.StdEncoding.DecodeString(certBase64)
	assert.NoError(t, errDecodeCert)
	key, errDecodeKey := base64.StdEncoding.DecodeString(keyBase64)
	assert.NoError(t, errDecodeKey)
	cert1 := &types.Certificate{
		Identifier:     "dev.cert.aha-0",
		Main:           "dev.cert.aha",
		Domains:        types.Domains{"dev.cert.aha", "dev2.cert.aha"},
		Certificate:    cert,
		Key:            key,
		ExpirationDate: time.Date(2035, time.April, 7, 00, 41, 15, 0, time.UTC),
	}
	certificateRaw := "\"Certificates\": [{\"Store\": \"default\",\"domain\": {\"main\": \"dev.cert.aha\", \"sans\": [\"dev.cert.aha\", \"dev2.cert.aha\"]}, \"certificate\": \"" + certBase64 + "\",\"key\": \"" + keyBase64 + "\"}]"
	tests := []struct {
		name       string
		mockFs     func(fs afero.Fs)
		want       *types.State
		wantErr    assert.ErrorAssertionFunc
		wantErrMsg string
	}{
		{
			name: "success",
			mockFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(filepath.Base(acmePath), 0755)
				_ = afero.WriteFile(fs, acmePath, []byte("{\"test\": {"+accountRaw+", "+certificateRaw+"}}"), 0644)
			},
			want:    &types.State{Account: account, Certificates: types.Certificates{cert1}},
			wantErr: assert.NoError,
		},
		{
			name: "failAcmeFileNotFound",
			mockFs: func(fs afero.Fs) {

			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "failed to read Traefik ACME file",
		},
		{
			name: "failParseAcmeFile",
			mockFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(filepath.Base(acmePath), 0755)
				_ = afero.WriteFile(fs, acmePath, []byte("{]"), 0644)
			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "failed to unmarshal Traefik ACME file: invalid character",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			ctx.Fs = fs
			tt.mockFs(fs)
			got, err := ReadTraefikData(ctx, acmePath)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadTraefikData(%v, %v)", ctx, acmePath)) {
				return
			}
			if err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}
			assert.Equalf(t, tt.want, got, "ReadTraefikData(%v, %v)", ctx, acmePath)
		})
	}
}

func TestMigrateTraefik(t *testing.T) {
	ctx := context.TestContext(nil)
	acmePath := "/app/letsencrypt/acme.json"

	tests := []struct {
		name       string
		mockFs     func(fs afero.Fs)
		want       *types.State
		wantErr    assert.ErrorAssertionFunc
		wantErrMsg string
	}{
		{
			name: "success",
			mockFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(filepath.Base(acmePath), 0755)
				_ = afero.WriteFile(fs, acmePath, []byte("{}"), 0644)
			},
			want:    &types.State{Account: &acme.Account{}, Certificates: types.Certificates{}},
			wantErr: assert.NoError,
		},
		{
			name:       "failReadAcmeFile",
			mockFs:     func(fs afero.Fs) {},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "failed to read Traefik ACME file: open /app/letsencrypt/acme.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			ctx.Fs = fs
			tt.mockFs(fs)
			got, err := MigrateTraefik(ctx, acmePath)
			if !tt.wantErr(t, err, fmt.Sprintf("MigrateTraefik(%v, %v)", ctx, acmePath)) {
				return
			}
			if err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}
			assert.Equalf(t, tt.want, got, "MigrateTraefik(%v, %v)", ctx, acmePath)
		})
	}
}
