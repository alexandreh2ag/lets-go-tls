package cli

import (
	"fmt"
	agentConfig "github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
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
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte(""), 0644)
	want := agentConfig.DefaultConfig()
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
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/foo.yml", path), []byte(""), 0644)
	want := agentConfig.DefaultConfig()
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

	want := agentConfig.DefaultConfig()
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
	_ = afero.WriteFile(fsFake, fmt.Sprintf("%s/config.yml", path), []byte("storages: [wrong]"), 0644)
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
	storagesStr := "storages: [{id: foo, type: fs, config: {path: '/app'}}]"
	requestersStr := "requesters: [{id: foo, type: static, config: {domains: [[foo.com]]}}]"
	managerStr := "manager: {address: http://127.0.0.1, token: secret}"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(globalStr+"\n"+stateStr+"\n"+storagesStr+"\n"+requestersStr+"\n"+managerStr), 0644)
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
	stateStr := "state: {type: fs, config: {path: '/app/acme.json'}}"
	managerStr := "manager: {address: http://127.0.0.1, token: secret}"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(stateStr+"\n"+managerStr), 0644)
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

func TestGetRootPreRunEFn_FailCreateStorages(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.WorkingDir = "/app"
	cmd := GetRootCmd(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	path := ctx.WorkingDir
	_ = ctx.Fs.Mkdir(path, 0775)
	storagesStr := "storages: [{id: foo, type: fs, agentConfig: {}}]"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(storagesStr), 0644)
	viper.Reset()
	viper.SetFs(ctx.Fs)
	_ = cmd.Execute()
	err := GetRootPreRunEFn(ctx, false)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create storages")
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
	providersStr := "providers: [{id: foo, type: static, storages: [foo], config: {domains: [foo.com]}}]"
	_ = afero.WriteFile(ctx.Fs, fmt.Sprintf("%s/config.yml", path), []byte(providersStr+"\n"), 0644)
	viper.Reset()
	viper.SetFs(ctx.Fs)
	err := GetRootPreRunEFn(ctx, true)(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file is not valid")
}
