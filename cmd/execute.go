package cmd

import (
	"fmt"
	"github.com/destanyinside/mefi/pkg/k8s"
	"github.com/destanyinside/mefi/pkg/log"
	"github.com/destanyinside/mefi/pkg/structs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const appName = "mefi"

var (

	// RootCmd is our main entry point, launching runE()
	RootCmd = &cobra.Command{
		Use:           appName,
		Short:         "Create envoy configuration from several k8s cluster",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRun:        bindConf,
		RunE:          runE,
	}
)

var (
	restcfg k8s.Interface
)

func runE(cmd *cobra.Command, args []string) error {

	logger, err := log.New(logLevel, logOutput)

	if err != nil {
		return fmt.Errorf("failed to create a logger: %v", err)
	}

	logger.Info(appName, " starting")

	config := &structs.Clusters{}

	err = viper.Unmarshal(config)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}

	return nil
}
