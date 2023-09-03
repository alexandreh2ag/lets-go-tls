package cli

import (
	"errors"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/cache"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	serverRequester "github.com/alexandreh2ag/lets-go-tls/apps/server/requester"
	appProm "github.com/alexandreh2ag/lets-go-tls/prometheus"
	"github.com/alexandreh2ag/lets-go-tls/requester"
	storageState "github.com/alexandreh2ag/lets-go-tls/storage/state"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"path"
)

const (
	Config   = "config"
	LogLevel = "level"
	Name     = types.NameServer
)

func GetRootCmd(ctx *appCtx.ServerContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:               Name,
		Short:             fmt.Sprintf("%s: server for centralized certificate manager", Name),
		PersistentPreRunE: GetRootPreRunEFn(ctx, true),
	}

	cmd.PersistentFlags().StringP(Config, "c", "", "Define config path")
	cmd.PersistentFlags().StringP(LogLevel, "l", "INFO", "Define log level")
	_ = viper.BindPFlag(Config, cmd.Flags().Lookup(Config))
	_ = viper.BindPFlag(LogLevel, cmd.Flags().Lookup(LogLevel))

	cmd.AddCommand(
		GetStartCmd(ctx),
		GetVersionCmd(),
	)

	return cmd
}

func GetRootPreRunEFn(ctx *appCtx.ServerContext, validateCfg bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		initConfig(ctx, cmd)

		if validateCfg {
			validate := validator.New()
			err = validate.Struct(ctx.Config)
			if err != nil {

				var validationErrors validator.ValidationErrors
				switch {
				case errors.As(err, &validationErrors):
					for _, validationError := range validationErrors {
						ctx.Logger.Error(fmt.Sprintf("%v", validationError))
					}
					return errors.New("configuration file is not valid")
				default:
					return err
				}
			}
		}
		serverRequester.Initialize()
		ctx.Requesters, err = requester.CreateRequesters(ctx, ctx.Config.Requesters)
		if err != nil {
			return fmt.Errorf("failed to create requesters: %v", err)
		}

		ctx.Cache, err = cache.CreateCache(ctx.Config.Cache)
		if err != nil {
			return fmt.Errorf("failed to create cache: %v", err)
		}

		ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

		ctx.StateStorage, err = storageState.CreateStorage(ctx, ctx.Config.State)
		if err != nil {
			return fmt.Errorf("failed to create state storage: %v", err)
		}

		logLevelFlagStr, _ := cmd.Flags().GetString(LogLevel)
		if logLevelFlagStr != "" {
			level := slog.LevelInfo
			err = level.UnmarshalText([]byte(logLevelFlagStr))
			if err != nil {
				return err
			}
			ctx.LogLevel.Set(level)
		}

		return nil
	}
}

func initConfig(ctx *appCtx.ServerContext, cmd *cobra.Command) {
	dir := ctx.WorkingDir

	viper.AddConfigPath(dir)
	viper.AutomaticEnv()
	viper.SetEnvPrefix(Name)
	viper.SetConfigName(Config)
	viper.SetConfigType("yaml")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		panic(err)
	}

	configPath := viper.GetString(Config)

	if configPath != "" {
		viper.SetConfigFile(configPath)
		configDir := path.Dir(configPath)
		if configDir != "." && configDir != dir {
			viper.AddConfigPath(configDir)
		}
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err)
	}

	err := viper.Unmarshal(ctx.Config)
	if err != nil {
		panic(fmt.Errorf("unable to decode into config struct, %v", err))
	}

}
