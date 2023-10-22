package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var (
	cfgFile   string
	selector  string
	resyncInt int
	logLevel  string
	logOutput string
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

	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "v", "info", "Log level")
	bindPFlag("log-level", "log-level")

	RootCmd.PersistentFlags().StringVarP(&logOutput, "log-output", "o", "stderr", "Log output")
	bindPFlag("log-output", "log-output")

	RootCmd.PersistentFlags().StringVarP(&selector, "filter", "f", "", "Label selector. Select only objects matching the label")
	bindPFlag("filter", "filter")

	RootCmd.PersistentFlags().IntVarP(&resyncInt, "resync-interval", "i", 900, "Full resync interval in seconds (0 to disable)")
	bindPFlag("resync-interval", "resync-interval")

}

// for whatever the reason, viper don't auto bind values from config file so we have to tell him
func bindConf(cmd *cobra.Command, args []string) {
	logLevel = viper.GetString("log-level")
	selector = viper.GetString("filter")
	resyncInt = viper.GetInt("resync-interval")
	logOutput = viper.GetString("log-output")
	cfgFile = viper.GetString("config")
}
