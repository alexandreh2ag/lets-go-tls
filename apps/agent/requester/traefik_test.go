package requester

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	mockHttp "github.com/alexandreh2ag/lets-go-tls/mocks/http"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	traefikConfigDynamic "github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikTypes "github.com/traefik/traefik/v3/pkg/types"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"
)

func Test_traefik_ID(t *testing.T) {
	want := "foo"
	instance := &traefik{id: want}
	assert.Equalf(t, want, instance.ID(), "ID()")
}

func Test_createTraefikV2Provider(t *testing.T) {
	ctx := context.TestContext(nil)
	want := &traefik{id: "foo", logger: ctx.Logger}
	want.addresses = []string{"127.0.0.1:80", "127.0.0.1:81"}
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
				Config: map[string]interface{}{"addresses": []string{"127.0.0.1:80", "127.0.0.1:81"}},
			},
			want:    want,
			wantErr: false,
		},
		{
			name: "FailDecodeCfg",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"addresses": "127.0.0.1:80"},
			},
			want:        want,
			wantErr:     true,
			errContains: "'addresses' source data must be an array or slice, got string",
		},
		{
			name: "FailValidateCfg",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"addresses": []string{}},
			},
			want:        want,
			wantErr:     true,
			errContains: "Key: 'ConfigTraefik.Addresses' Error:Field validation for 'Addresses' failed on the 'min' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createTraefikV2Provider(ctx, tt.cfg)

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

func Test_traefik_FetchIntance(t *testing.T) {
	ctrl := gomock.NewController(t)
	a := &traefik{}
	want := []types.DomainRequest{
		{Domains: types.Domains{types.Domain("foo.com")}},
		{Domains: types.Domains{types.Domain("bar.com")}},
	}
	tests := []struct {
		name    string
		want    []types.DomainRequest
		wantErr assert.ErrorAssertionFunc

		funcMock func(clientHttp *mockHttp.MockClient)
	}{
		{
			name:    "Success",
			want:    want,
			wantErr: assert.NoError,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusOK)
				body, _ := json.Marshal([]traefikConfigDynamic.Router{{Rule: "Host(`foo.com`)", TLS: &traefikConfigDynamic.RouterTLSConfig{}}, {Rule: "Host(`bar.com`)", TLS: &traefikConfigDynamic.RouterTLSConfig{}}})
				resp.SetBody(body)
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
		},
		{
			name:    "FailDoRequest",
			want:    []types.DomainRequest{},
			wantErr: assert.Error,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail"))
			},
		},
		{
			name:    "FailStatusCode",
			want:    []types.DomainRequest{},
			wantErr: assert.Error,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusInternalServerError)
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
		},
		{
			name:    "FailUnmarshal",
			want:    []types.DomainRequest{},
			wantErr: assert.Error,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusOK)
				resp.SetBody([]byte("{]"))
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientHttp := mockHttp.NewMockClient(ctrl)
			tt.funcMock(clientHttp)
			a.httpClient = clientHttp

			got, err := a.FetchInstance("127.0.0.1:80")
			if !tt.wantErr(t, err, fmt.Sprintf("FetchInstance(127.0.0.1:80)")) {
				return
			}
			requests := []types.DomainRequest{}
			for _, request := range got {
				requests = append(requests, *request)
			}
			assert.Equalf(t, tt.want, requests, "FetchInstance(127.0.0.1:80)")
		})
	}
}

func Test_traefik_Fetch(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	a := &traefik{addresses: []string{"127.0.0.1:80", "127.0.0.1:81"}, logger: ctx.Logger}
	want1 := []types.DomainRequest{
		{Requester: a, Domains: types.Domains{types.Domain("foo.com")}},
	}
	want2 := []types.DomainRequest{
		{Requester: a, Domains: types.Domains{types.Domain("bar.com")}},
	}
	want := append(want1, want2...)
	tests := []struct {
		name     string
		funcMock func(clientHttp *mockHttp.MockClient)
		want     []types.DomainRequest
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:    "Success",
			want:    want,
			wantErr: assert.NoError,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp1 := fasthttp.Response{}
				resp1.SetStatusCode(http.StatusOK)
				body, _ := json.Marshal([]traefikConfigDynamic.Router{
					{Rule: "Host(`foo.com`)", TLS: &traefikConfigDynamic.RouterTLSConfig{}},
					{Rule: "Host(`127.0.0.1`)", TLS: &traefikConfigDynamic.RouterTLSConfig{}},
				})
				resp1.SetBody(body)

				resp2 := fasthttp.Response{}
				resp2.SetStatusCode(http.StatusOK)
				body, _ = json.Marshal([]traefikConfigDynamic.Router{{Rule: "Host(`bar.com`)", TLS: &traefikConfigDynamic.RouterTLSConfig{}}})
				resp2.SetBody(body)

				gomock.InOrder(
					clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp1).Return(nil),
					clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp2).Return(nil),
				)

			},
		},
		{
			name:    "Fail",
			want:    want1,
			wantErr: assert.Error,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp1 := fasthttp.Response{}
				resp1.SetStatusCode(http.StatusOK)
				body, _ := json.Marshal([]traefikConfigDynamic.Router{{Rule: "Host(`foo.com`)", TLS: &traefikConfigDynamic.RouterTLSConfig{}}})
				resp1.SetBody(body)
				gomock.InOrder(
					clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp1).Return(nil),
					clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail")),
				)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientHttp := mockHttp.NewMockClient(ctrl)
			tt.funcMock(clientHttp)
			a.httpClient = clientHttp
			got, err := a.Fetch()
			if !tt.wantErr(t, err, fmt.Sprintf("Fetch()")) {
				return
			}

			requests := []types.DomainRequest{}
			for _, request := range got {
				requests = append(requests, *request)
			}
			assert.ElementsMatchf(t, tt.want, requests, "Fetch()")
		})
	}
}

func Test_traefik_FormatRouters(t *testing.T) {

	tests := []struct {
		name    string
		routers []traefikConfigDynamic.Router
		want    []types.DomainRequest
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			routers: []traefikConfigDynamic.Router{
				{Rule: "Host(`notssl.foo.com`)"},
				{Rule: "Host(`foo.com`)", TLS: &traefikConfigDynamic.RouterTLSConfig{}},
				{Rule: "Path(`/foo`)", TLS: &traefikConfigDynamic.RouterTLSConfig{}},
				{Rule: "Host(`a.foo.com`) || (Host(`b.foo.com`) && Path(`/bar`))", TLS: &traefikConfigDynamic.RouterTLSConfig{}},
				{
					Rule: "Host(`sub.bar.com`)",
					TLS: &traefikConfigDynamic.RouterTLSConfig{
						Domains: []traefikTypes.Domain{
							{Main: "bar.com", SANs: []string{"*.bar.com"}},
							{Main: "sub.bar.com", SANs: []string{"*.sub.bar.com"}},
						},
					},
				},
			},
			want: []types.DomainRequest{
				{Domains: types.Domains{types.Domain("foo.com")}},
				{Domains: types.Domains{types.Domain("a.foo.com"), types.Domain("b.foo.com")}},
				{Domains: types.Domains{types.Domain("bar.com"), types.Domain("*.bar.com")}},
				{Domains: types.Domains{types.Domain("sub.bar.com"), types.Domain("*.sub.bar.com")}},
			},
			wantErr: assert.NoError,
		},
		{
			name: "FailParseRule",
			routers: []traefikConfigDynamic.Router{
				{Rule: "Host(`foo.com", TLS: &traefikConfigDynamic.RouterTLSConfig{}},
			},
			want:    []types.DomainRequest{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := &traefik{
				id: "foo",
			}
			got, err := instance.FormatRouters(tt.routers)
			if !tt.wantErr(t, err, fmt.Sprintf("FormatRouters(%v)", tt.routers)) {
				return
			}

			requests := []types.DomainRequest{}
			for _, request := range got {
				requests = append(requests, *request)
			}
			assert.Equalf(t, tt.want, requests, "FormatRouters(%v)", tt.routers)
		})
	}
}
