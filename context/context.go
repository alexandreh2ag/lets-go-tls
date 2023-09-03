package context

import (
	"github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/alexandreh2ag/lets-go-tls/prometheus"
	"github.com/spf13/afero"
	"github.com/valyala/fasthttp"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

type Context interface {
	GetFS() afero.Fs

	GetLogger() *slog.Logger
	GetLogLevel() *slog.LevelVar

	GetMetricsRegister() prometheus.Registry

	GetWorkingDir() string

	GetHttpClient() http.Client

	Cancel()
	Done() <-chan bool
	Signal() chan os.Signal
}

var _ Context = &BaseContext{}

type BaseContext struct {
	Logger     *slog.Logger
	LogLevel   *slog.LevelVar
	WorkingDir string
	Fs         afero.Fs
	sigs       chan os.Signal
	done       chan bool

	HttpClient http.Client

	MetricsRegister prometheus.Registry
}

func (c *BaseContext) GetFS() afero.Fs {
	return c.Fs
}

func (c *BaseContext) GetLogger() *slog.Logger {
	return c.Logger
}

func (c *BaseContext) GetLogLevel() *slog.LevelVar {
	return c.LogLevel
}

func (c *BaseContext) GetMetricsRegister() prometheus.Registry {
	return c.MetricsRegister
}

func (c *BaseContext) GetWorkingDir() string {
	return c.WorkingDir
}

func (c *BaseContext) Cancel() {
	c.done <- true
}

func (c *BaseContext) Done() <-chan bool {
	return c.done
}

func (c *BaseContext) Signal() chan os.Signal {
	return c.sigs
}

func (c *BaseContext) GetHttpClient() http.Client {
	return c.HttpClient
}

func DefaultContext() *BaseContext {
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	return &BaseContext{
		Logger:     slog.New(slog.NewTextHandler(os.Stdout, opts)),
		LogLevel:   level,
		WorkingDir: workingDir,
		Fs:         afero.NewOsFs(),
		done:       make(chan bool),
		sigs:       sigs,
		HttpClient: &fasthttp.Client{},
	}
}

func TestContext(logBuffer io.Writer) *BaseContext {
	if logBuffer == nil {
		logBuffer = io.Discard
	}
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	return &BaseContext{
		Logger:   slog.New(slog.NewTextHandler(logBuffer, opts)),
		LogLevel: level,
		Fs:       afero.NewMemMapFs(),
		done:     make(chan bool),
		sigs:     sigs,
	}
}
