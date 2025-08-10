package nginx

import (
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_isValidDomainName(t *testing.T) {

	tests := []struct {
		name   string
		domain string
		want   bool
	}{
		{
			name:   "valid domain name",
			domain: "example.com",
			want:   true,
		},
		{
			name:   "valid sub domain name",
			domain: "test.example.com",
			want:   true,
		},
		{
			name:   "valid domain name with digit",
			domain: "exampl3.co1",
			want:   true,
		},
		{
			name:   "valid domain name with hyphen",
			domain: "test-example.com",
			want:   true,
		},
		{
			name:   "valid sub domain name with hyphen",
			domain: "test.test-example.com",
			want:   true,
		},
		{
			name:   "valid sub domain name with hyphen and digit",
			domain: "1.test-example.com",
			want:   true,
		},
		{
			name:   "invalid domain name",
			domain: "(example).com",
			want:   false,
		},
		{
			name:   "invalid sub domain name",
			domain: "(test).example.com",
			want:   false,
		},
		{
			name:   "invalid wildcard domain name",
			domain: "*.example.com",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidDomainName(tt.domain); got != tt.want {
				t.Errorf("isValidDomainName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidDomainNames(t *testing.T) {

	tests := []struct {
		name    string
		domains types.Domains
		want    bool
	}{
		{
			name:    "valid domain names",
			domains: types.Domains{"example.com", "foo.example.com"},
			want:    true,
		},
		{
			name:    "invalid domain names",
			domains: types.Domains{"example.com", "(foo).example.com", "foo.example.com"},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isValidDomainNames(tt.domains), "isValidDomainNames(%v)", tt.domains)
		})
	}
}

func TestParseConfig(t *testing.T) {
	ctx := context.TestContext(nil)
	type args struct {
		cfgPath string
	}
	tests := []struct {
		name    string
		cfgPath string
		args    args
		want    VhostConfigs
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Success",
			cfgPath: "./fixtures/nginx_valid/nginx.conf",
			want: VhostConfigs{
				{ServerName: types.Domains{"example.com"}, KeyPath: "/etc/ssl/example.com.key", CertPath: "/etc/ssl/example.com.crt"},
				{ServerName: types.Domains{"foo.example.com", "bar.example.com"}, KeyPath: "/etc/ssl/foo.example.com.key", CertPath: "/etc/ssl/foo.example.com.crt"},
			},
			wantErr: assert.NoError,
		},
		{
			name:    "FailedReadConfig",
			cfgPath: "./fixtures/wrong.conf",
			want:    VhostConfigs{},
			wantErr: assert.Error,
		},
		{
			name:    "FailedParseConfig",
			cfgPath: "./fixtures/nginx_invalid/nginx.conf",
			want:    VhostConfigs{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfig(ctx.Logger, tt.cfgPath)
			tt.wantErr(t, err, "ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.want, got, "ParseConfig() got = %v, want %v", got, tt.want)
		})
	}
}
