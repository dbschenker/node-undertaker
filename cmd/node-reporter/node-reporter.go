package main

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodereporter"

	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-reporter/flags"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "node-reporter",
	Short: "Node reporter reports node status as lease",
	Long: "Node reporter polls custom http endpoint and reports its status as lease" +
		"Please use `node-reporter --help` to get possible options",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Node-reporter starting")
		err := flags.ValidateRootFlags()
		if err != nil {
			return err
		}

		return nodereporter.Execute()
	},
}
