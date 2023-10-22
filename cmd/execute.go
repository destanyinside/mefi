package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/destanyinside/mefi/pkg/controller"
	"github.com/destanyinside/mefi/pkg/discovery"
	"github.com/destanyinside/mefi/pkg/k8s"
	"github.com/destanyinside/mefi/pkg/log"
	"github.com/destanyinside/mefi/pkg/structs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/dynamic"
	"os"
	"os/signal"
	"syscall"
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

	var factory *controller.Factory
	var discover *discovery.Discover
	var discoveries []*discovery.Discover

	for _, i := range config.Clusters {
		ca, err := base64.StdEncoding.DecodeString(i.Ca)
		restcfg, err = k8s.New(i.Url, ca, i.Token)
		if err != nil {
			return fmt.Errorf("failed to create a client: %v", err)
		}

		clt := dynamic.NewForConfigOrDie(restcfg.GetRestConfig())
		k8sCli := &structs.K8sClient{ClusterName: i.Name, DInf: clt}
		factory = controller.NewFactory(logger, selector, resyncInt)
		discover = discovery.New(logger, factory, selector, k8sCli).Start()
		discoveries = append(discoveries, discover)
	}

	logger.Info(appName, " started")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm

	for _, i := range discoveries {
		logger.Info(appName, " stopping")
		i.Stop()
	}
	if err != nil {
		return err
	}
	logger.Info(appName, " stopped")

	return nil
}

// Execute adds all child commands to the root command and sets their flags.
func Execute() error {
	return RootCmd.Execute()
}
