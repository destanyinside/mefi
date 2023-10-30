package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var (
	cfgFile                   string
	restConfigQps             float32
	restConfigBurst           int
	remoteLabelSelector       string
	localLabelSelector        string
	originalNameLabelSelector string
	mefiNamespace             string
	resyncInt                 int
	logLevel                  string
	logOutput                 string
)

func bindPFlag(key string, cmd string) {
	if err := viper.BindPFlag(key, RootCmd.PersistentFlags().Lookup(cmd)); err != nil {
		log.Fatal("Failed to bind cli argument:", err)
	}
}

func init() {
	cobra.OnInitialize(loadConfigFile)
	RootCmd.AddCommand(versionCmd)

	defaultCfg := "/etc/gce/" + appName + ".yaml"
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultCfg, "Configuration file")

	RootCmd.PersistentFlags().Float32VarP(&restConfigQps, "k8s-rest-config-qps", "q", 5, "Indicates the maximum QPS to the master from this client")
	bindPFlag("k8s-rest-config-qps", "k8s-rest-config-qps")

	RootCmd.PersistentFlags().IntVarP(&restConfigBurst, "k8s-rest-config-burst", "b", 10, "Maximum burst for throttle to the master from this client")
	bindPFlag("k8s-rest-config-burst", "k8s-rest-config-burst")

	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "v", "info", "Log level")
	bindPFlag("log-level", "log-level")

	RootCmd.PersistentFlags().StringVarP(&logOutput, "log-output", "o", "stderr", "Log output")
	bindPFlag("log-output", "log-output")

	RootCmd.PersistentFlags().StringVarP(&remoteLabelSelector, "remote-filter", "", "isMefiRemote", "Label name for remote cluster. Select only objects matching the label")
	bindPFlag("remote-filter", "remote-filter")

	RootCmd.PersistentFlags().StringVarP(&localLabelSelector, "local-filter", "", "isMefiLocal", "Label name for local cluster. Select only objects matching the label")
	bindPFlag("local-filter", "local-filter")

	RootCmd.PersistentFlags().StringVarP(&originalNameLabelSelector, "original-filter", "", "isMefiOriginalName", "Label name for original endpoint name")
	bindPFlag("original-filter", "original-filter")

	RootCmd.PersistentFlags().StringVarP(&mefiNamespace, "mefi-namespace", "n", "mefi-system", "System namespace for "+appName)
	bindPFlag("mefi-namespace", "mefi-namespace")

	RootCmd.PersistentFlags().IntVarP(&resyncInt, "resync-interval", "i", 900, "Full resync interval in seconds (0 to disable)")
	bindPFlag("resync-interval", "resync-interval")

}

func bindConf(cmd *cobra.Command, args []string) {
	logLevel = viper.GetString("log-level")
	restConfigQps = float32(viper.GetFloat64("k8s-rest-config-qps"))
	restConfigBurst = viper.GetInt("k8s-rest-config-burst")
	remoteLabelSelector = viper.GetString("remote-filter")
	localLabelSelector = viper.GetString("local-filter")
	originalNameLabelSelector = viper.GetString("original-filter")
	mefiNamespace = viper.GetString("mefi-namespace")
	resyncInt = viper.GetInt("resync-interval")
	logOutput = viper.GetString("log-output")
	cfgFile = viper.GetString("config")
}
