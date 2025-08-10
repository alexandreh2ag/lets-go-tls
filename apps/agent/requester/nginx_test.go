package requester

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_nginx_ID(t *testing.T) {
	n := nginx{id: NginxKey}
	assert.Equalf(t, NginxKey, n.ID(), "ID()")
}

func Test_nginx_Fetch(t *testing.T) {
	ctx := context.TestContext(nil)
	tests := []struct {
		name    string
		cfg     ConfigNginx
		want    []*types.DomainRequest
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			cfg:  ConfigNginx{NginxCfgPath: "../../../nginx/fixtures/nginx_valid/nginx.conf"},
			want: []*types.DomainRequest{
				{Domains: types.Domains{"example.com"}},
				{Domains: types.Domains{"foo.example.com", "bar.example.com"}},
			},
			wantErr: assert.NoError,
		},
		{
			name:    "FailedParseConfig",
			cfg:     ConfigNginx{NginxCfgPath: "./wrong/nginx.conf"},
			want:    []*types.DomainRequest{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := nginx{
				id:     "nginx",
				cfg:    tt.cfg,
				logger: ctx.GetLogger(),
			}
			for _, request := range tt.want {
				request.Requester = n
			}
			got, err := n.Fetch()
			if !tt.wantErr(t, err, fmt.Sprintf("Fetch()")) {
				return
			}
			assert.Equalf(t, tt.want, got, "Fetch()")
		})
	}
}

func Test_createNginxProvider(t *testing.T) {
	ctx := context.TestContext(nil)
	want := &nginx{id: "foo", logger: ctx.GetLogger(), cfg: ConfigNginx{NginxCfgPath: "/etc/nginx/nginx.conf"}}
	tests := []struct {
		name        string
		cfg         config.RequesterConfig
		want        types.Requester
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"nginx_cfg_path": "/etc/nginx/nginx.conf"},
			},
			want:    want,
			wantErr: false,
		},
		{
			name: "FailDecodeCfg",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"nginx_cfg_path": 1},
			},
			want:        want,
			wantErr:     true,
			errContains: "expected type 'string'",
		},
		{
			name: "FailValidateCfg",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{},
			},
			want:        want,
			wantErr:     true,
			errContains: "Error:Field validation for 'NginxCfgPath' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createNginxProvider(ctx, tt.cfg)
			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
