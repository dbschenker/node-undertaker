package main

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

const (
	LogLevelFlag = "log-level"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "node-undertaker",
	Short: "Node undertaker terminates kubernetes nodes that are unhealthy",
	Long: "Node undertaker terminates kubernetes nodes that are unhealthy" +
		"Please use `node-undertaker --help` to get possible options",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Please use `node-undertaker --help` to get possible options")
		err := validateRootFlags()
		if err != nil {
			return err
		}

		return nodeundertaker.Execute()
	},
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {

	cobra.OnInitialize(initConfig)
	//flags
	rootCmd.PersistentFlags().String(LogLevelFlag, "info", "Log level [panic|fatal|error|warn|warning|info|debug|trace]. Default: 'info'. Can be set using LOG_LEVEL env variable")
	err := viper.BindPFlag(LogLevelFlag, rootCmd.PersistentFlags().Lookup(LogLevelFlag))
	if err != nil {
		panic(err)
	}

}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func validateRootFlags() error {
	viper.GetString(LogLevelFlag)
	_, err := log.ParseLevel(viper.GetString(LogLevelFlag))
	if err != nil {
		return err
	}
	return nil
}
