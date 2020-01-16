package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/superfly/flyctl/docstrings"
	"os"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/helpers"
	"github.com/superfly/flyctl/flyctl"
)

func newConfigCommand() *Command {

	configStrings := docstrings.Get("config")

	cmd := &Command{
		Command: &cobra.Command{
			Use:   configStrings.Usage,
			Short: configStrings.Short,
			Long:  configStrings.Long,
		},
	}

	configDisplayStrings := docstrings.Get("config.display")
	BuildCommand(cmd, runDisplayConfig, configDisplayStrings.Usage, configDisplayStrings.Short, configDisplayStrings.Long, true, os.Stdout, requireAppName)

	configSaveStrings := docstrings.Get("config.save")
	BuildCommand(cmd, runSaveConfig, configSaveStrings.Usage, configSaveStrings.Short, configSaveStrings.Long, true, os.Stdout, requireAppName)

	configValidateStrings := docstrings.Get("config.validate")
	BuildCommand(cmd, runValidateConfig, configValidateStrings.Usage, configValidateStrings.Short, configValidateStrings.Long, true, os.Stdout, requireAppName)

	return cmd
}

func runDisplayConfig(ctx *CmdContext) error {
	cfg, err := ctx.FlyClient.GetConfig(ctx.AppName)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(cfg.Definition)

	return nil
}

func runSaveConfig(ctx *CmdContext) error {
	if ctx.AppConfig == nil {
		ctx.AppConfig = flyctl.NewAppConfig()
	}
	ctx.AppConfig.AppName = ctx.AppName

	serverCfg, err := ctx.FlyClient.GetConfig(ctx.AppName)
	if err != nil {
		return err
	}

	ctx.AppConfig.Definition = serverCfg.Definition

	return writeAppConfig(ctx.ConfigFile, ctx.AppConfig)
}

func runValidateConfig(ctx *CmdContext) error {
	if ctx.AppConfig == nil {
		return errors.New("App config file not found")
	}

	fmt.Println("Validating", ctx.ConfigFile)

	serverCfg, err := ctx.FlyClient.ParseConfig(ctx.AppName, ctx.AppConfig.Definition)
	if err != nil {
		return err
	}

	if serverCfg.Valid {
		fmt.Println(aurora.Green("✓").String(), "Configuration is valid")
		return nil
	}

	printAppConfigErrors(*serverCfg)

	return errors.New("App configuration is not valid")
}

func printAppConfigErrors(cfg api.AppConfig) {
	fmt.Println()
	for _, error := range cfg.Errors {
		fmt.Println("   ", aurora.Red("✘").String(), error)
	}
	fmt.Println()
}

func printAppConfigServices(prefix string, cfg api.AppConfig) {
	if len(cfg.Services) > 0 {
		fmt.Println(prefix, "Services")
		for _, svc := range cfg.Services {
			fmt.Println(prefix, "  ", svc.Description)
		}
	}
}

func writeAppConfig(path string, appConfig *flyctl.AppConfig) error {
	if !confirmFileOverwrite(path) {
		return nil
	}

	if err := appConfig.WriteToFile(path); err != nil {
		return err
	}

	fmt.Println("Wrote config file", helpers.PathRelativeToCWD(path))

	return nil
}
