package cli

import (
	"fmt"
	serverConfig "github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func Test_initConfig_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	fsFake := ctx.Fs
	viper.Reset()
	viper.SetFs(fsFake)
	path := ctx.WorkingDir
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("requesters: []"), 0644)
	want := serverConfig.DefaultConfig()
	want.Requesters = []config.RequesterConfig{}
	initConfig(ctx, cmd)
	assert.Equal(t, &want, ctx.Config)
}

func Test_initConfig_SuccessWithConfigFlag(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	fsFake := ctx.Fs
	viper.Reset()
	viper.SetFs(fsFake)
	path := "/foo"
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/foo.yml", path), []byte("requesters: []"), 0644)
	want := serverConfig.DefaultConfig()
	want.Requesters = []config.RequesterConfig{}
	viper.Set(Config, fmt.Sprintf("%s/foo.yml", path))
	initConfig(ctx, cmd)
	assert.Equal(t, &want, ctx.Config)
}

func Test_initConfig_FailReadConfig(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	fsFake := ctx.Fs
	viper.Reset()
	viper.SetFs(fsFake)

	want := serverConfig.DefaultConfig()
	initConfig(ctx, cmd)
	assert.Equal(t, &want, ctx.Config)
}

func Test_initConfig_FailUnmarshal(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	fsFake := ctx.Fs
	viper.Reset()
	viper.SetFs(fsFake)
	path := ctx.WorkingDir
	_ = fsFake.Mkdir(path, 0775)
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("requesters: [wrong]"), 0644)
	defer func() {
		if r := recover(); r != nil {
			assert.True(t, true)
		} else {
			t.Errorf("initConfig should have panicked")
		}
	}()
	initConfig(ctx, cmd)
}

func TestGetRootPreRunEFn_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	path := ctx.WorkingDir
	_ = ctx.Fs.Mkdir(path, 0775)
	globalStr := ""
	stateStr := "state: {type: fs, config: {path: '/app/acme.json'}}"
	requestersStr := "requesters: [{id: foo, type: static, config: {domains: [[foo.com]]}}]"
	acmeStr := "acme: {email: dev@foo.com}"
	jwtStr := "jwt: {key: secret}"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(globalStr+"\n"+stateStr+"\n"+requestersStr+"\n"+acmeStr+"\n"+jwtStr), 0644)
	viper.Reset()
	viper.SetFs(ctx.Fs)
	err := GetRootPreRunEFn(ctx, true)(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "LevelVar(INFO)", ctx.LogLevel.String())
}

func TestGetRootPreRunEFn_SuccessLogLevelFlag(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	path := ctx.WorkingDir
	_ = ctx.Fs.Mkdir(path, 0775)
	globalStr := ""
	stateStr := "state: {type: fs, config: {path: '/app/acme.json'}}"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(globalStr+"\n"+stateStr), 0644)
	viper.Reset()
	viper.SetFs(ctx.Fs)
	cmd.SetArgs([]string{
		"--" + LogLevel, "ERROR"},
	)
	_ = cmd.Execute()
	err := GetRootPreRunEFn(ctx, false)(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "LevelVar(ERROR)", ctx.LogLevel.String())
}

func TestGetRootPreRunEFn_FailCreateRequesters(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	path := ctx.WorkingDir
	_ = ctx.Fs.Mkdir(path, 0775)
	requestersStr := "requesters: [{id: foo, type: static, config: {domains: []}}]"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(requestersStr+"\n"), 0644)
	viper.Reset()
	viper.SetFs(ctx.Fs)
	_ = cmd.Execute()
	err := GetRootPreRunEFn(ctx, false)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create requesters")
}

func TestGetRootPreRunEFn_FailCreateCache(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	path := ctx.WorkingDir
	_ = ctx.Fs.Mkdir(path, 0775)
	globalStr := ""
	stateStr := "state: {type: fs, config: {path: '/app/acme.json'}}"
	cacheStr := "cache: {type: wrong, config: {}}"
	requestersStr := "requesters: [{id: foo, type: static, config: {domains: [[foo.com]]}}]"
	acmeStr := "acme: {email: dev@foo.com}"
	jwtStr := "jwt: {key: secret}"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(globalStr+"\n"+stateStr+"\n"+cacheStr+"\n"+requestersStr+"\n"+acmeStr+"\n"+jwtStr), 0644)
	viper.Reset()
	viper.SetFs(ctx.Fs)
	err := GetRootPreRunEFn(ctx, true)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create cache")
}

func TestGetRootPreRunEFn_FailLogLevelFlagInvalid(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	path := ctx.WorkingDir
	_ = ctx.Fs.Mkdir(path, 0775)
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(""), 0644)
	viper.Reset()
	viper.SetFs(ctx.Fs)
	cmd.SetArgs([]string{
		"--" + LogLevel, "WRONG"},
	)
	_ = cmd.Execute()
	err := GetRootPreRunEFn(ctx, false)(cmd, []string{})
	assert.Error(t, err)
	assert.Equal(t, "LevelVar(INFO)", ctx.LogLevel.String())
}

func TestGetRootPreRunEFn_FailValidator(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	path := ctx.WorkingDir
	_ = ctx.Fs.Mkdir(path, 0775)
	providersStr := "providers: [{id: foo, type: static, config: {domains: [foo.com]}}]"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(providersStr+"\n"), 0644)
	viper.Reset()
	viper.SetFs(ctx.Fs)
	err := GetRootPreRunEFn(ctx, true)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file is not valid")
}
