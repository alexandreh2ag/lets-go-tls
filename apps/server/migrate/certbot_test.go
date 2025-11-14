package migrate

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/registration"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReadCertbotAccount(t *testing.T) {
	ctx := context.TestContext(nil)
	dirPath := "/app/letsencrypt/account"

	regrValid := "{\"body\": {}, \"uri\": \"http://server/acme/name/directory/azerty\"}"
	keyValid := "{\"n\": \"9pbYlW2rfzxM9BnhKnWXWQvOaSofpbgpLnDZqBb5UJuOexknabPVVjwAJCXyn004Q4J7mjkNsc6xK42Pd8l_2Z1FODh7NJjw_TDBIMn44WQqRF1tZe29WIZBvaW1g7VXkJ4oI93osUf4sqQ-24tUOXYktAVirITk4tIgfiKAV-KI28Cye2CkCmMZtWi_yKoySo-dVLUj1BfZ3BBeDRNVEr2VcNtq6qqZ54phtblcRJPaJ2ZRr77hZJN-DU5Q1IttnJqf1OJkndlRbAdgZ5uHH_k3w-6MlobD1YZPT6bSsSoZMx6YRt14PVWLhuGu2NB_l3JFbGHF2kw3JsQUX5695w\", \"e\": \"AQAB\", \"d\": \"Cwc2CoDMIbWdr3EU2-REb4wSoCttHOk-lkAfy9-zKMK8PY8jbxvG18N5MiSsVCmM8Z_9TSluDlyzGcKL_wx49k-NT_VetTx90qUCjifaRKJJLaOMB_n49OOUPxoRIyHSF7qrCueax4rzeXHVCdaSXruE6rQu9I5A-k_xehKq8MMvaDBSY_Zwbmyd3QkeXSZq08CRL0nmKizwKamoEbxNdeHZDXyYtJKUaFRpS311Kr74OuC6C7jQRSIz4GlAN2DhgLMnJaxFZkvIWknkprTHRzlmCj6RNmRF9OVFlBhJTJdJjFAyfMLpdwzOki-Q5YdbxWrcA-7NJgHvvcZf1k_leQ\", \"p\": \"_dc0cSGPqeGwXUs1zVuOQc7sJODIjlSZ-Jt8GJ1geQBU8wLCFYkw-Fa6EfJUxk8qhgHpjgelLCv1m8x5fxPeUVafWSug7XqnwSvDpaThS1SNbzOPsyYYSI7RBuVLJz5y949fxziPawf0-Ip6Vv1zKaGKbz-gZFyJyZ8wxRbIaqU\", \"q\": \"-K_ZgKYrFHX96ZnjsVUyTQ0HZUAJGvNMaicA8sa_EgiDGtiDLO0yn6btV4uaVMiXy_5GRPikaHO0xD3DYilkGk0WGnswmhy24NyV7H5NDqNyT6BNFDldEe0_8N3WFs1ppY-tzWQgnSbazIuLq5wbR2TcyzAXJM4bXTjIXD0RvJs\", \"dp\": \"X_-QXKyVmWi5-z5jVYunjCnGUtgT4QgLxuJ-lwyFnJ1Mgy7q2Zwtwe2CpyDfgLQV3-I_QrCPfdoOI-K7nByWemppDx8Z2FYHtV1ou29UqNmMN57nXJgunNouIQ58UpYigo9daIlya8bxlUFBrT5i3D60jQgiD3KTxYWHuzG3j-U\", \"dq\": \"vqvVT9YX4bA15g2fji-MSZdbvD3EvW0dgaY2C-5mQSVVoBexO5kB33bVMvJOegLyn-1tYyRgqdVNG8lfLLwfjCreb59UPFUXwoBeUtzjp_2Zf4OglYkF2uDUzZDIvOLxxYbL81Z5ywQnbfzwMbuXDr4-q2tL02RThr4qoM4R81E\", \"qi\": \"LcHVIWfnbkq668LDsqJk48fsNwfJUjmxsvUUrf9YYVrjU2v4r6SV0DYb-UWjwccc5KZpnNY1sZGnbpSMoob3fqDhPb50cLMgLfj7DmVxR9hGIRnNgAYAf2f3mUfAyvjVwJGIKGoBS8vsnXbDmzZRGntKBSRIFvRYkCkVDEoXi8Q\", \"kty\": \"RSA\"}"
	keyBase64 := "MIIEowIBAAKCAQEA9pbYlW2rfzxM9BnhKnWXWQvOaSofpbgpLnDZqBb5UJuOexknabPVVjwAJCXyn004Q4J7mjkNsc6xK42Pd8l/2Z1FODh7NJjw/TDBIMn44WQqRF1tZe29WIZBvaW1g7VXkJ4oI93osUf4sqQ+24tUOXYktAVirITk4tIgfiKAV+KI28Cye2CkCmMZtWi/yKoySo+dVLUj1BfZ3BBeDRNVEr2VcNtq6qqZ54phtblcRJPaJ2ZRr77hZJN+DU5Q1IttnJqf1OJkndlRbAdgZ5uHH/k3w+6MlobD1YZPT6bSsSoZMx6YRt14PVWLhuGu2NB/l3JFbGHF2kw3JsQUX5695wIDAQABAoIBAAsHNgqAzCG1na9xFNvkRG+MEqArbRzpPpZAH8vfsyjCvD2PI28bxtfDeTIkrFQpjPGf/U0pbg5csxnCi/8MePZPjU/1XrU8fdKlAo4n2kSiSS2jjAf5+PTjlD8aESMh0he6qwrnmseK83lx1QnWkl67hOq0LvSOQPpP8XoSqvDDL2gwUmP2cG5snd0JHl0matPAkS9J5ios8CmpqBG8TXXh2Q18mLSSlGhUaUt9dSq++Drgugu40EUiM+BpQDdg4YCzJyWsRWZLyFpJ5Ka0x0c5Zgo+kTZkRfTlRZQYSUyXSYxQMnzC6XcMzpIvkOWHW8Vq3APuzSYB773GX9ZP5XkCgYEA/dc0cSGPqeGwXUs1zVuOQc7sJODIjlSZ+Jt8GJ1geQBU8wLCFYkw+Fa6EfJUxk8qhgHpjgelLCv1m8x5fxPeUVafWSug7XqnwSvDpaThS1SNbzOPsyYYSI7RBuVLJz5y949fxziPawf0+Ip6Vv1zKaGKbz+gZFyJyZ8wxRbIaqUCgYEA+K/ZgKYrFHX96ZnjsVUyTQ0HZUAJGvNMaicA8sa/EgiDGtiDLO0yn6btV4uaVMiXy/5GRPikaHO0xD3DYilkGk0WGnswmhy24NyV7H5NDqNyT6BNFDldEe0/8N3WFs1ppY+tzWQgnSbazIuLq5wbR2TcyzAXJM4bXTjIXD0RvJsCgYBf/5BcrJWZaLn7PmNVi6eMKcZS2BPhCAvG4n6XDIWcnUyDLurZnC3B7YKnIN+AtBXf4j9CsI992g4j4rucHJZ6amkPHxnYVge1XWi7b1So2Yw3nudcmC6c2i4hDnxSliKCj11oiXJrxvGVQUGtPmLcPrSNCCIPcpPFhYe7MbeP5QKBgQC+q9VP1hfhsDXmDZ+OL4xJl1u8PcS9bR2BpjYL7mZBJVWgF7E7mQHfdtUy8k56AvKf7W1jJGCp1U0byV8svB+MKt5vn1Q8VRfCgF5S3OOn/Zl/g6CViQXa4NTNkMi84vHFhsvzVnnLBCdt/PAxu5cOvj6ra0vTZFOGviqgzhHzUQKBgC3B1SFn525KuuvCw7KiZOPH7DcHyVI5sbL1FK3/WGFa41Nr+K+kldA2G/lFo8HHHOSmaZzWNbGRp26UjKKG936g4T2+dHCzIC34+w5lcUfYRiEZzYAGAH9n95lHwMr41cCRiChqAUvL7J12w5s2URp7SgUkSBb0WJApFQxKF4vE"
	key, errDecode := base64.StdEncoding.DecodeString(keyBase64)
	assert.NoError(t, errDecode)
	account := &acme.Account{Email: "", Registration: &registration.Resource{URI: "http://server/acme/name/directory/azerty"}, Key: key}
	tests := []struct {
		name       string
		want       *acme.Account
		mockFs     func(fs afero.Fs)
		wantErr    assert.ErrorAssertionFunc
		wantErrMsg string
	}{
		{
			name: "success",
			mockFs: func(fs afero.Fs) {
				accountDir := filepath.Join(dirPath, "server", "acme", "name", "directory", "azerty")
				_ = fs.MkdirAll(accountDir, 0755)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "regr.json"), []byte(regrValid), 0644)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "private_key.json"), []byte(keyValid), 0644)
			},
			want:    account,
			wantErr: assert.NoError,
		},
		{
			name: "failDirectoryNotExist",
			mockFs: func(fs afero.Fs) {
			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "directory /app/letsencrypt/account does not exist",
		},
		{
			name: "failedFoundAccount",
			mockFs: func(fs afero.Fs) {
				accountDir := filepath.Join(dirPath, "server", "acme", "name", "directory", "azerty")
				_ = fs.MkdirAll(accountDir, 0755)
			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "no account not found",
		},
		{
			name: "failedReadKey",
			mockFs: func(fs afero.Fs) {
				accountDir := filepath.Join(dirPath, "server", "acme", "name", "directory", "azerty")
				_ = fs.MkdirAll(accountDir, 0755)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "regr.json"), []byte(""), 0644)
			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "open /app/letsencrypt/account/server/acme/name/directory/azerty/private_key.json: file does not exist",
		},
		{
			name: "failedParseRegr",
			mockFs: func(fs afero.Fs) {
				accountDir := filepath.Join(dirPath, "server", "acme", "name", "directory", "azerty")
				_ = fs.MkdirAll(accountDir, 0755)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "regr.json"), []byte("{]"), 0644)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "private_key.json"), []byte(""), 0644)
			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "failed to unmarshal acme regr.json file: invalid character",
		},
		{
			name: "failedParseKey",
			mockFs: func(fs afero.Fs) {
				accountDir := filepath.Join(dirPath, "server", "acme", "name", "directory", "azerty")
				_ = fs.MkdirAll(accountDir, 0755)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "regr.json"), []byte(regrValid), 0644)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "private_key.json"), []byte("[}"), 0644)
			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "failed to unmarshal acme private_key.json file: invalid character",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			ctx.Fs = fs
			tt.mockFs(fs)
			got, err := ReadCertbotAccount(ctx, dirPath)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadCertbotAccount(ctx, %v)", dirPath)) {
				return
			}
			if err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}

			if got != nil {
				assert.Equal(t, tt.want.Email, got.Email)
				assert.Equal(t, tt.want.Key, got.Key)
				if !reflect.DeepEqual(got.Registration, tt.want.Registration) {
					t.Errorf("ReadCertbotAccount() got.Registration = %v, want.Registration %v", got.Registration, tt.want.Registration)
				}
			}
		})
	}
}

func TestReadCertbotCertificates(t *testing.T) {
	ctx := context.TestContext(nil)
	dirPath := "/app/letsencrypt/live"

	certValid := []byte(`-----BEGIN CERTIFICATE-----
MIIB9jCCAZygAwIBAgIQKF1wjbLoPHF01uW456SlMjAKBggqhkjOPQQDAjA8MRQw
EgYDVQQKEwtSZWZsZXQgRGV2bDEkMCIGA1UEAxMbUmVmbGV0IERldmwgSW50ZXJt
ZWRpYXRlIENBMB4XDTI1MDQwNjEyNDAxNVoXDTM1MDQwNzAwNDExNVowADBZMBMG
ByqGSM49AgEGCCqGSM49AwEHA0IABLGbA6BkSSTPsBFNjEw/exybt1znWoz4W1xj
+aGBtHt1C4c/e10FS5kIGvZoGZaX1VVhiPNngofAVnJGpcP31d2jgbswgbgwDgYD
VR0PAQH/BAQDAgeAMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAdBgNV
HQ4EFgQU3c2aBQQIe/C8gt91uxcluINGzfQwHwYDVR0jBBgwFoAU3ftzTjtS2b9E
imhSL7VF6f4A/XYwKQYDVR0RAQH/BB8wHYIMZGV2LmNlcnQuYWhhgg1kZXYyLmNl
cnQuYWhhMBwGDCsGAQQBgqRkxihAAQQMMAoCAQYEA2RldgQAMAoGCCqGSM49BAMC
A0gAMEUCIGqG0Q2Dy2+cxMCeDcrMGQmYfWeX/qmPEgyYEUsdj20hAiEAl22fqIB4
ObtzJCuTEQKhid013R38oIGGoLCxrPO4BVY=
-----END CERTIFICATE-----
`)
	keyValid := []byte(`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgmN9Y4LzULrSs114E
EquocxGnp4hY0Tb/HQnfI2xnyyehRANCAASxmwOgZEkkz7ARTYxMP3scm7dc51qM
+FtcY/mhgbR7dQuHP3tdBUuZCBr2aBmWl9VVYYjzZ4KHwFZyRqXD99Xd
-----END PRIVATE KEY-----
`)
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
	tests := []struct {
		name       string
		mockFs     func(fs afero.Fs)
		want       types.Certificates
		wantErr    assert.ErrorAssertionFunc
		wantErrMsg string
	}{
		{
			name: "successNoCertificates",
			mockFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(dirPath, 0755)
			},
			want:    types.Certificates{},
			wantErr: assert.NoError,
		},
		{
			name: "success",
			mockFs: func(fs afero.Fs) {
				dirCert1 := filepath.Join(dirPath, "dev.cert.aha")
				_ = fs.MkdirAll(dirPath, 0755)
				_ = fs.Mkdir(dirCert1, 0755)
				_ = afero.WriteFile(fs, filepath.Join(dirCert1, "fullchain.pem"), certValid, 0644)
				_ = afero.WriteFile(fs, filepath.Join(dirCert1, "privkey.pem"), keyValid, 0644)
			},
			want:    types.Certificates{cert1},
			wantErr: assert.NoError,
		},
		{
			name: "failDirectoryNotExist",
			mockFs: func(fs afero.Fs) {
			},
			want:       types.Certificates{},
			wantErr:    assert.Error,
			wantErrMsg: "directory /app/letsencrypt/live does not exist",
		},
		{
			name: "failReadCert",
			mockFs: func(fs afero.Fs) {
				dirCert1 := filepath.Join(dirPath, "dev.cert.aha")
				_ = fs.MkdirAll(dirPath, 0755)
				_ = fs.Mkdir(dirCert1, 0755)
			},
			want:       types.Certificates{},
			wantErr:    assert.Error,
			wantErrMsg: "open /app/letsencrypt/live/dev.cert.aha/fullchain.pem: file does not exist",
		},
		{
			name: "failReadKey",
			mockFs: func(fs afero.Fs) {
				dirCert1 := filepath.Join(dirPath, "dev.cert.aha")
				_ = fs.MkdirAll(dirPath, 0755)
				_ = fs.Mkdir(dirCert1, 0755)
				_ = afero.WriteFile(fs, filepath.Join(dirCert1, "fullchain.pem"), []byte("wrong"), 0644)
				_ = afero.WriteFile(fs, filepath.Join(dirCert1, "privkey.pem"), keyValid, 0644)
			},
			want:       types.Certificates{},
			wantErr:    assert.Error,
			wantErrMsg: "failed to decode certificate pem /app/letsencrypt/live/dev.cert.aha/fullchain.pem: failed to decode cert",
		},
		{
			name: "failParseCert",
			mockFs: func(fs afero.Fs) {
				dirCert1 := filepath.Join(dirPath, "dev.cert.aha")
				_ = fs.MkdirAll(dirPath, 0755)
				_ = fs.Mkdir(dirCert1, 0755)
				_ = afero.WriteFile(fs, filepath.Join(dirCert1, "fullchain.pem"), certValid, 0644)
			},
			want:       types.Certificates{},
			wantErr:    assert.Error,
			wantErrMsg: "open /app/letsencrypt/live/dev.cert.aha/privkey.pem: file does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			ctx.Fs = fs
			tt.mockFs(fs)
			got, err := ReadCertbotCertificates(ctx, dirPath)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadCertbotCertificates(%v, %v)", ctx, dirPath)) {
				return
			}
			if err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}

			assert.Equalf(t, tt.want, got, "ReadCertbotCertificates(%v, %v)", ctx, dirPath)
		})
	}
}

func TestMigrateCertbot(t *testing.T) {
	ctx := context.TestContext(nil)
	dirPath := "/app/letsencrypt"

	regrValid := "{\"body\": {}, \"uri\": \"http://server/acme/name/directory/azerty\"}"
	keyValid := "{\"n\": \"9pbYlW2rfzxM9BnhKnWXWQvOaSofpbgpLnDZqBb5UJuOexknabPVVjwAJCXyn004Q4J7mjkNsc6xK42Pd8l_2Z1FODh7NJjw_TDBIMn44WQqRF1tZe29WIZBvaW1g7VXkJ4oI93osUf4sqQ-24tUOXYktAVirITk4tIgfiKAV-KI28Cye2CkCmMZtWi_yKoySo-dVLUj1BfZ3BBeDRNVEr2VcNtq6qqZ54phtblcRJPaJ2ZRr77hZJN-DU5Q1IttnJqf1OJkndlRbAdgZ5uHH_k3w-6MlobD1YZPT6bSsSoZMx6YRt14PVWLhuGu2NB_l3JFbGHF2kw3JsQUX5695w\", \"e\": \"AQAB\", \"d\": \"Cwc2CoDMIbWdr3EU2-REb4wSoCttHOk-lkAfy9-zKMK8PY8jbxvG18N5MiSsVCmM8Z_9TSluDlyzGcKL_wx49k-NT_VetTx90qUCjifaRKJJLaOMB_n49OOUPxoRIyHSF7qrCueax4rzeXHVCdaSXruE6rQu9I5A-k_xehKq8MMvaDBSY_Zwbmyd3QkeXSZq08CRL0nmKizwKamoEbxNdeHZDXyYtJKUaFRpS311Kr74OuC6C7jQRSIz4GlAN2DhgLMnJaxFZkvIWknkprTHRzlmCj6RNmRF9OVFlBhJTJdJjFAyfMLpdwzOki-Q5YdbxWrcA-7NJgHvvcZf1k_leQ\", \"p\": \"_dc0cSGPqeGwXUs1zVuOQc7sJODIjlSZ-Jt8GJ1geQBU8wLCFYkw-Fa6EfJUxk8qhgHpjgelLCv1m8x5fxPeUVafWSug7XqnwSvDpaThS1SNbzOPsyYYSI7RBuVLJz5y949fxziPawf0-Ip6Vv1zKaGKbz-gZFyJyZ8wxRbIaqU\", \"q\": \"-K_ZgKYrFHX96ZnjsVUyTQ0HZUAJGvNMaicA8sa_EgiDGtiDLO0yn6btV4uaVMiXy_5GRPikaHO0xD3DYilkGk0WGnswmhy24NyV7H5NDqNyT6BNFDldEe0_8N3WFs1ppY-tzWQgnSbazIuLq5wbR2TcyzAXJM4bXTjIXD0RvJs\", \"dp\": \"X_-QXKyVmWi5-z5jVYunjCnGUtgT4QgLxuJ-lwyFnJ1Mgy7q2Zwtwe2CpyDfgLQV3-I_QrCPfdoOI-K7nByWemppDx8Z2FYHtV1ou29UqNmMN57nXJgunNouIQ58UpYigo9daIlya8bxlUFBrT5i3D60jQgiD3KTxYWHuzG3j-U\", \"dq\": \"vqvVT9YX4bA15g2fji-MSZdbvD3EvW0dgaY2C-5mQSVVoBexO5kB33bVMvJOegLyn-1tYyRgqdVNG8lfLLwfjCreb59UPFUXwoBeUtzjp_2Zf4OglYkF2uDUzZDIvOLxxYbL81Z5ywQnbfzwMbuXDr4-q2tL02RThr4qoM4R81E\", \"qi\": \"LcHVIWfnbkq668LDsqJk48fsNwfJUjmxsvUUrf9YYVrjU2v4r6SV0DYb-UWjwccc5KZpnNY1sZGnbpSMoob3fqDhPb50cLMgLfj7DmVxR9hGIRnNgAYAf2f3mUfAyvjVwJGIKGoBS8vsnXbDmzZRGntKBSRIFvRYkCkVDEoXi8Q\", \"kty\": \"RSA\"}"
	keyBase64 := "MIIEowIBAAKCAQEA9pbYlW2rfzxM9BnhKnWXWQvOaSofpbgpLnDZqBb5UJuOexknabPVVjwAJCXyn004Q4J7mjkNsc6xK42Pd8l/2Z1FODh7NJjw/TDBIMn44WQqRF1tZe29WIZBvaW1g7VXkJ4oI93osUf4sqQ+24tUOXYktAVirITk4tIgfiKAV+KI28Cye2CkCmMZtWi/yKoySo+dVLUj1BfZ3BBeDRNVEr2VcNtq6qqZ54phtblcRJPaJ2ZRr77hZJN+DU5Q1IttnJqf1OJkndlRbAdgZ5uHH/k3w+6MlobD1YZPT6bSsSoZMx6YRt14PVWLhuGu2NB/l3JFbGHF2kw3JsQUX5695wIDAQABAoIBAAsHNgqAzCG1na9xFNvkRG+MEqArbRzpPpZAH8vfsyjCvD2PI28bxtfDeTIkrFQpjPGf/U0pbg5csxnCi/8MePZPjU/1XrU8fdKlAo4n2kSiSS2jjAf5+PTjlD8aESMh0he6qwrnmseK83lx1QnWkl67hOq0LvSOQPpP8XoSqvDDL2gwUmP2cG5snd0JHl0matPAkS9J5ios8CmpqBG8TXXh2Q18mLSSlGhUaUt9dSq++Drgugu40EUiM+BpQDdg4YCzJyWsRWZLyFpJ5Ka0x0c5Zgo+kTZkRfTlRZQYSUyXSYxQMnzC6XcMzpIvkOWHW8Vq3APuzSYB773GX9ZP5XkCgYEA/dc0cSGPqeGwXUs1zVuOQc7sJODIjlSZ+Jt8GJ1geQBU8wLCFYkw+Fa6EfJUxk8qhgHpjgelLCv1m8x5fxPeUVafWSug7XqnwSvDpaThS1SNbzOPsyYYSI7RBuVLJz5y949fxziPawf0+Ip6Vv1zKaGKbz+gZFyJyZ8wxRbIaqUCgYEA+K/ZgKYrFHX96ZnjsVUyTQ0HZUAJGvNMaicA8sa/EgiDGtiDLO0yn6btV4uaVMiXy/5GRPikaHO0xD3DYilkGk0WGnswmhy24NyV7H5NDqNyT6BNFDldEe0/8N3WFs1ppY+tzWQgnSbazIuLq5wbR2TcyzAXJM4bXTjIXD0RvJsCgYBf/5BcrJWZaLn7PmNVi6eMKcZS2BPhCAvG4n6XDIWcnUyDLurZnC3B7YKnIN+AtBXf4j9CsI992g4j4rucHJZ6amkPHxnYVge1XWi7b1So2Yw3nudcmC6c2i4hDnxSliKCj11oiXJrxvGVQUGtPmLcPrSNCCIPcpPFhYe7MbeP5QKBgQC+q9VP1hfhsDXmDZ+OL4xJl1u8PcS9bR2BpjYL7mZBJVWgF7E7mQHfdtUy8k56AvKf7W1jJGCp1U0byV8svB+MKt5vn1Q8VRfCgF5S3OOn/Zl/g6CViQXa4NTNkMi84vHFhsvzVnnLBCdt/PAxu5cOvj6ra0vTZFOGviqgzhHzUQKBgC3B1SFn525KuuvCw7KiZOPH7DcHyVI5sbL1FK3/WGFa41Nr+K+kldA2G/lFo8HHHOSmaZzWNbGRp26UjKKG936g4T2+dHCzIC34+w5lcUfYRiEZzYAGAH9n95lHwMr41cCRiChqAUvL7J12w5s2URp7SgUkSBb0WJApFQxKF4vE"
	key, errDecode := base64.StdEncoding.DecodeString(keyBase64)
	assert.NoError(t, errDecode)
	account := &acme.Account{Email: "", Registration: &registration.Resource{URI: "http://server/acme/name/directory/azerty"}, Key: key}

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
				accountDir := filepath.Join(dirPath, "accounts")
				_ = fs.MkdirAll(dirPath, 0755)
				_ = fs.Mkdir(filepath.Join(dirPath, "live"), 0755)
				_ = fs.Mkdir(filepath.Join(dirPath, "accounts"), 0755)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "regr.json"), []byte(regrValid), 0644)
				_ = afero.WriteFile(fs, filepath.Join(accountDir, "private_key.json"), []byte(keyValid), 0644)
			},
			want:    &types.State{Account: account, Certificates: types.Certificates{}},
			wantErr: assert.NoError,
		},
		{
			name: "failReadCertificate",
			mockFs: func(fs afero.Fs) {

			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "directory /app/letsencrypt/live does not exist",
		},
		{
			name: "failReadAccount",
			mockFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(dirPath, 0755)
				_ = fs.Mkdir(filepath.Join(dirPath, "live"), 0755)
			},
			want:       nil,
			wantErr:    assert.Error,
			wantErrMsg: "directory /app/letsencrypt/accounts does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			ctx.Fs = fs
			tt.mockFs(fs)
			got, err := MigrateCertbot(ctx, dirPath)
			if !tt.wantErr(t, err, fmt.Sprintf("MigrateCertbot(%v, %v)", ctx, dirPath)) {
				return
			}

			if err != nil {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}

			assert.Equalf(t, tt.want, got, "MigrateCertbot(%v, %v)", ctx, dirPath)
		})
	}
}
