package main

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "node-undertaker",
	Short: "Node undertaker terminates kubernetes nodes that are unhealthy",
	Long: "Node undertaker terminates kubernetes nodes that are unhealthy" +
		"Please use `node-undertaker --help` to get possible options",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Node-undertaker starting")
		err := flags.ValidateRootFlags()
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
	err := flags.SetupFlags(rootCmd)
	if err != nil {
		panic(err)
	}
}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
