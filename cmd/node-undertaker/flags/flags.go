package flags

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	LogLevelFlag              = "log-level"
	CloudProviderFlag         = "cloud-provider"
	DrainDelayFlag            = "drain-delay"
	CloudTerminationDelayFlag = "cloud-termination-delay"
	PortFlag                  = "port"
	NodeInitialThresholdFlag  = "node-initial-threshold"
	NamespaceFlag             = "namespace"
	LeaseLockNameFlag         = "lease-lock-name"
	LeaseLockNamespaceFlag    = "lease-lock-namespace"
)

func SetupFlags(cmd *cobra.Command) error {
	cmd.PersistentFlags().String(LogLevelFlag, "info", "Log level [panic|fatal|error|warn|warning|info|debug|trace]. Default: 'info'. Can be set using LOG_LEVEL env variable")
	err := viper.BindPFlag(LogLevelFlag, cmd.PersistentFlags().Lookup(LogLevelFlag))
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
	cmd.PersistentFlags().String(NamespaceFlag, "", "Namespace containing leases. Default: '' - which is the same namespace node-undertaker runs. Can be set using NAMESPACE env variable")
	err = viper.BindPFlag(NamespaceFlag, cmd.PersistentFlags().Lookup(NamespaceFlag))
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

	return nil
}

func ValidateRootFlags() error {
	viper.GetString(LogLevelFlag)
	_, err := log.ParseLevel(viper.GetString(LogLevelFlag))
	if err != nil {
		return err
	}

	return nil
}
