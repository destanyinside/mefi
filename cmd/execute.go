package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/destanyinside/mefi/pkg/event"
	"github.com/destanyinside/mefi/pkg/k8s"
	"github.com/destanyinside/mefi/pkg/localApplier"
	"github.com/destanyinside/mefi/pkg/localWatcher"
	"github.com/destanyinside/mefi/pkg/log"
	"github.com/destanyinside/mefi/pkg/remoteWatcher"
	"github.com/destanyinside/mefi/pkg/structs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func runE(cmd *cobra.Command, args []string) error {

	logger, err := log.NewLogger(logLevel, logOutput)

	if err != nil {
		return fmt.Errorf("failed to create a logger: %v", err)
	}

	logger.Infof("%s starting", appName)

	config := &structs.Clusters{}

	err = viper.Unmarshal(config)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}

	var rWatch *remoteWatcher.Watcher
	var rWatchers []*remoteWatcher.Watcher
	var apply *localApplier.Applier
	var lWatch *localWatcher.Watcher

	eventNotifier := event.New()

	for _, i := range config.Clusters {
		ca, err := base64.StdEncoding.DecodeString(i.Ca)
		restCfg := k8s.New(i.Url, ca, i.Token, logger, restConfigQps, restConfigBurst)
		if err != nil {
			return fmt.Errorf("failed to create a client: %v", err)
		}
		clientSet := &structs.K8sClient{ClusterName: i.Name, ClientSet: restCfg}
		rWatch = remoteWatcher.NewWatcher(logger, remoteLabelSelector, localLabelSelector, originalNameLabelSelector, resyncInt, clientSet, eventNotifier)
		rWatch.Start()
		if i.Type == structs.Local {
			apply = localApplier.NewApplier(logger, mefiNamespace, clientSet, eventNotifier)
			lWatch = localWatcher.NewWatcher(logger, localLabelSelector, originalNameLabelSelector, resyncInt, mefiNamespace, clientSet, eventNotifier)
			apply.Start()
			lWatch.Start()
		}
		rWatchers = append(rWatchers, rWatch)
	}

	logger.Infof("%s started", appName)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm

	for _, i := range rWatchers {
		logger.Infof("watchers stopping")
		i.Stop()
	}
	apply.Stop()
	logger.Infof("local creator stopping")
	lWatch.Stop()
	logger.Infof("local watcher stopping")
	if err != nil {
		return err
	}
	logger.Infof("stopped %s", appName)

	return nil
}

func Execute() error {
	return RootCmd.Execute()
}
