package flags

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	LogLevelFlag  = "log-level"
	LogFormatFlag = "log-format"
	NamespaceFlag = "namespace"
	NodeNameFlag  = "node-name"
	UrlFlag       = "url"
	TimeoutFlag   = "timeout"
	FrequencyFlag = "frequency"
	LeaseTimeFlag = "lease-time"

	LogFormatJson = "json"
	LogFormatText = "text"
)

func SetupFlags(cmd *cobra.Command) error {
	cmd.PersistentFlags().String(LogLevelFlag, "info", "Log level [panic|fatal|error|warn|warning|info|debug|trace]. Default: 'info'. Can be set using LOG_LEVEL env variable")
	err := viper.BindPFlag(LogLevelFlag, cmd.PersistentFlags().Lookup(LogLevelFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().String(LogFormatFlag, "text", "Log format [text|json]. Default: 'text'. Can be set using LOG_FORMAT env variable")
	err = viper.BindPFlag(LogFormatFlag, cmd.PersistentFlags().Lookup(LogFormatFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().String(NamespaceFlag, "", "Namespace containing leases. Default: '' - which is the same namespace node-reporter runs. Can be set using NAMESPACE env variable")
	err = viper.BindPFlag(NamespaceFlag, cmd.PersistentFlags().Lookup(NamespaceFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().String(NodeNameFlag, "", "Node name and node lease name. Can be set using NODE_NAME env variable")
	err = viper.BindPFlag(NodeNameFlag, cmd.PersistentFlags().Lookup(NodeNameFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().String(UrlFlag, "", "Url to check node healthiness")
	err = viper.BindPFlag(UrlFlag, cmd.PersistentFlags().Lookup(UrlFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().Int(TimeoutFlag, 10, "Timeout (in seconds) for checking healthiness of the node. Can be set using TIMEOUT env variable.")
	err = viper.BindPFlag(TimeoutFlag, cmd.PersistentFlags().Lookup(TimeoutFlag))
	cmd.PersistentFlags().Int(FrequencyFlag, 20, "How long (in seconds) takes between lease updates (default 20). Can be set using FREQUENCY env variable.")
	err = viper.BindPFlag(FrequencyFlag, cmd.PersistentFlags().Lookup(FrequencyFlag))
	cmd.PersistentFlags().Int(LeaseTimeFlag, 90, "Node healthiness lease time in seconds (default 90). Can be set using LEASE_TIME env variable.")
	err = viper.BindPFlag(LeaseTimeFlag, cmd.PersistentFlags().Lookup(LeaseTimeFlag))

	if err != nil {
		return err
	}
	return nil
}

func ValidateRootFlags() error {
	viper.GetString(LogLevelFlag)
	_, err := log.ParseLevel(viper.GetString(LogLevelFlag))
	if err != nil {
		return err
	}

	format := viper.GetString(LogFormatFlag)
	if format != LogFormatJson && format != LogFormatText {
		return fmt.Errorf("unknown log format: %s", format)
	}

	nodeName := viper.GetString(NodeNameFlag)
	if nodeName == "" {
		return errors.New("node name not set. Please set it using --node-name flag or NODE_NAME env variable")
	}

	return nil
}
