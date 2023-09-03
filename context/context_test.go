package context

import (
	appProm "github.com/alexandreh2ag/lets-go-tls/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestDefaultContext(t *testing.T) {

	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	fs := afero.NewOsFs()
	workingDir, err := os.Getwd()
	assert.NoError(t, err)
	want := &BaseContext{
		WorkingDir: workingDir,
		Logger:     logger,
		LogLevel:   level,
		Fs:         fs,
		HttpClient: &fasthttp.Client{},
	}
	got := DefaultContext()
	assert.NotNil(t, got.done)
	got.done = nil
	got.sigs = nil
	assert.Equal(t, want, got)
}

func TestTestContext(t *testing.T) {

	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	logger := slog.New(slog.NewTextHandler(io.Discard, opts))
	fs := afero.NewMemMapFs()
	want := &BaseContext{
		Logger:   logger,
		LogLevel: level,
		Fs:       fs,
	}
	got := TestContext(nil)
	assert.NotNil(t, got.done)
	got.done = nil
	got.sigs = nil
	assert.Equal(t, want, got)
}

func TestTestContext_WithLogBuffer(t *testing.T) {

	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	logger := slog.New(slog.NewTextHandler(io.Discard, opts))
	fs := afero.NewMemMapFs()
	want := &BaseContext{
		Logger:   logger,
		LogLevel: level,
		Fs:       fs,
	}
	got := TestContext(io.Discard)
	assert.NotNil(t, got.done)
	got.done = nil
	got.sigs = nil
	assert.Equal(t, want, got)
}

func TestContext_Cancel(t *testing.T) {
	ctx := &BaseContext{}
	ctx.done = make(chan bool)
	running := true
	go func() {
		select {
		case <-ctx.done:
			running = false
		}
	}()
	ctx.Cancel()
	assert.Equal(t, false, running)
}

func TestContext_Done(t *testing.T) {
	ctx := &BaseContext{}
	ctx.done = make(chan bool)
	running := true
	go func() {
		select {
		case <-ctx.Done():
			running = false
		}
	}()
	ctx.done <- true
	assert.Equal(t, false, running)
}

func TestContext_Signal(t *testing.T) {
	ctx := &BaseContext{}
	ctx.sigs = make(chan os.Signal, 1)
	signal.Notify(ctx.sigs, syscall.SIGINT, syscall.SIGTERM)
	running := true
	go func() {
		select {
		case <-ctx.Signal():
			running = false
		}
	}()
	ctx.Signal() <- syscall.SIGINT
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, running)
}

func TestBaseContext_GetFS(t *testing.T) {
	fs := afero.NewMemMapFs()
	c := &BaseContext{
		Fs: fs,
	}
	assert.Equalf(t, fs, c.GetFS(), "GetFS()")
}

func TestBaseContext_GetLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false, Level: &slog.LevelVar{}}))
	c := &BaseContext{
		Logger: logger,
	}
	assert.Equalf(t, logger, c.GetLogger(), "GetLogger()")
}

func TestBaseContext_GetLogLevel(t *testing.T) {
	logLevel := &slog.LevelVar{}
	logLevel.Set(slog.LevelInfo)
	c := &BaseContext{
		LogLevel: logLevel,
	}
	assert.Equalf(t, logLevel, c.GetLogLevel(), "GetLogLevel()")
}

func TestBaseContext_GetMetricsRegister(t *testing.T) {
	register := appProm.NewRegistry("foo", prometheus.NewRegistry())
	c := &BaseContext{
		MetricsRegister: register,
	}
	assert.Equalf(t, register, c.GetMetricsRegister(), "GetMetricsRegister()")
}

func TestBaseContext_GetWorkingDir(t *testing.T) {
	workingDir := "/app"
	c := &BaseContext{
		WorkingDir: workingDir,
	}
	assert.Equalf(t, workingDir, c.GetWorkingDir(), "GetWorkingDir()")
}

func TestBaseContext_GetHttpClient(t *testing.T) {
	client := &fasthttp.Client{}
	c := &BaseContext{
		HttpClient: client,
	}
	assert.Equalf(t, client, c.GetHttpClient(), "GetHttpClient()")
}
