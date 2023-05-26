package flags

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	LogLevelFlag              = "log-level"
	LogFormatFlag             = "log-format"
	CloudProviderFlag         = "cloud-provider"
	InitialDelayFlag          = "initial-delay"
	DrainDelayFlag            = "drain-delay"
	CloudTerminationDelayFlag = "cloud-termination-delay"
	PortFlag                  = "port"
	NodeInitialThresholdFlag  = "node-initial-threshold"
	NodeLeaseNamespaceFlag    = "node-lease-namespace"
	NamespaceFlag             = "namespace"
	LeaseLockNameFlag         = "lease-lock-name"
	LeaseLockNamespaceFlag    = "lease-lock-namespace"
	LogFormatJson             = "json"
	LogFormatText             = "text"
	NodeSelectorFlag          = "node-selector"
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
	cmd.PersistentFlags().String(CloudProviderFlag, "aws", "Cloud provider name. Default: 'aws'. Possible values: aws,kwok,kind. Can be set using CLOUD_PROVIDER env variable")
	err = viper.BindPFlag(CloudProviderFlag, cmd.PersistentFlags().Lookup(CloudProviderFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().Int(DrainDelayFlag, 300, "Drain unhealthy node after number of seconds after observed unhealthy (env: DRAIN_DELAY)")
	err = viper.BindPFlag(DrainDelayFlag, cmd.PersistentFlags().Lookup(DrainDelayFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().Int(CloudTerminationDelayFlag, 300, "Terminate unhealthy node after number of seconds after starting drain (env: CLOUD_TERMINATION_DELAY)")
	err = viper.BindPFlag(CloudTerminationDelayFlag, cmd.PersistentFlags().Lookup(CloudTerminationDelayFlag))
	cmd.PersistentFlags().Int(NodeInitialThresholdFlag, 120, "Node is skipped until this number of seconds passes since creation (env: NODE_INITIAL_THRESHOLD)")
	err = viper.BindPFlag(NodeInitialThresholdFlag, cmd.PersistentFlags().Lookup(NodeInitialThresholdFlag))
	if err != nil {
		return err
	}

	if err != nil {
		panic(err)
	}
	cmd.PersistentFlags().Int(PortFlag, 8080, "Http port (used for observability). Can be set using PORT env variable")
	err = viper.BindPFlag(PortFlag, cmd.PersistentFlags().Lookup(PortFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().String(NamespaceFlag, "", "Namespace where events should be created. Default: '' - which is the same namespace node-undertaker runs. Can be set using NAMESPACE env variable")
	err = viper.BindPFlag(NamespaceFlag, cmd.PersistentFlags().Lookup(NamespaceFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().String(NodeLeaseNamespaceFlag, "kube-node-lease", "Namespace containing leases. Default: '' - which is the same namespace node-undertaker runs. Can be set using NODE_LEASE_NAMESPACE env variable")
	err = viper.BindPFlag(NodeLeaseNamespaceFlag, cmd.PersistentFlags().Lookup(NodeLeaseNamespaceFlag))
	if err != nil {
		return err
	}
	//lease
	cmd.PersistentFlags().String(LeaseLockNamespaceFlag, "", "Namespace containing leader election lease. Default: '' - which is the same namespace node-undertaker runs. Can be set using LEASE_LOCK_NAMESPACE env variable")
	err = viper.BindPFlag(LeaseLockNamespaceFlag, cmd.PersistentFlags().Lookup(LeaseLockNamespaceFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().String(LeaseLockNameFlag, "node-undertaker-leader-election", "Name of node-undertaker's leader election lease. Default: 'node-undertaker-leader-election'. Can be set using LEASE_LOCK_NAME env variable")
	err = viper.BindPFlag(LeaseLockNameFlag, cmd.PersistentFlags().Lookup(LeaseLockNameFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().Int(InitialDelayFlag, 60, "Initial delay from start of node-undertaker pod until starts handling node state changes. Default: '60'. Can be set using INITIAL_DELAY env variable")
	err = viper.BindPFlag(InitialDelayFlag, cmd.PersistentFlags().Lookup(InitialDelayFlag))
	if err != nil {
		return err
	}
	cmd.PersistentFlags().String(NodeSelectorFlag, "", "Label selector for nodes to watch. Default: ''. Can be set using NODE_SELECTOR env variable")
	err = viper.BindPFlag(NodeSelectorFlag, cmd.PersistentFlags().Lookup(NodeSelectorFlag))
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

	return nil
}
