package cli

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	mockTypesStorageState "github.com/alexandreh2ag/lets-go-tls/mocks/types/storage/state"
	appProm "github.com/alexandreh2ag/lets-go-tls/prometheus"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"syscall"
	"testing"
	"time"
)

func TestGetStartRunFn_SuccessOnlyListenHTTP(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.Config.HTTP.Listen = "127.0.0.1:0"
	viper.Reset()
	viper.SetFs(ctx.Fs)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := mockTypesStorageState.NewMockStorage(ctrl)
	storage.EXPECT().Load().AnyTimes().Return(&types.State{}, nil)
	storage.EXPECT().Save(gomock.Any()).AnyTimes().Return(nil)
	ctx.StateStorage = storage

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cmd := GetStartCmd(ctx)
	go func() {
		err := GetStartRunFn(ctx)(cmd, []string{})
		assert.NoError(t, err)
	}()
	time.Sleep(time.Millisecond * 100)
	ctx.Signal() <- syscall.SIGINT
}

func TestGetStartRunFn_SuccessListenHTTPS(t *testing.T) {
	ctx := context.TestContext(nil)
	ctx.Config.HTTP.Listen = "127.0.0.1:0"
	ctx.Config.HTTP.TLS.Enable = true
	ctx.Config.HTTP.TLS.Listen = "127.0.0.1:0"
	viper.Reset()
	viper.SetFs(ctx.Fs)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := mockTypesStorageState.NewMockStorage(ctrl)
	storage.EXPECT().Load().AnyTimes().Return(&types.State{}, nil)
	storage.EXPECT().Save(gomock.Any()).AnyTimes().Return(nil)
	ctx.StateStorage = storage

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cmd := GetStartCmd(ctx)
	go func() {
		err := GetStartRunFn(ctx)(cmd, []string{})
		assert.NoError(t, err)
	}()
	time.Sleep(time.Millisecond * 100)
	ctx.Signal() <- syscall.SIGINT
}
