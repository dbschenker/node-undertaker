package flags

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	LogLevelFlag      = "log-level"
	CloudProviderFlag = "cloud-provider"
	AwsSqsUrlFlag     = "aws-sqs-url"
	AwsRegionFlag     = "aws-region"
	DrainTimeoutFlag  = "drain-timeout"
	PortFlag          = "port"
)

func SetupFlags(cmd *cobra.Command) error {
	cmd.PersistentFlags().String(LogLevelFlag, "info", "Log level [panic|fatal|error|warn|warning|info|debug|trace]. Default: 'info'. Can be set using LOG_LEVEL env variable")
	err := viper.BindPFlag(LogLevelFlag, cmd.PersistentFlags().Lookup(LogLevelFlag))
	if err != nil {
		return (err)
	}
	cmd.PersistentFlags().String(CloudProviderFlag, "aws", "Cloud provider name. Default: 'aws'. Possible values: aws. Can be set using CLOUD_PROVIDER env variable")
	err = viper.BindPFlag(CloudProviderFlag, cmd.PersistentFlags().Lookup(CloudProviderFlag))
	if err != nil {
		panic(err)
	}
	cmd.PersistentFlags().String(AwsSqsUrlFlag, "", "Url for AWS SQS (in case cloud-provider=aws). Can be set using AWS_SQS_URL env variable")
	err = viper.BindPFlag(AwsSqsUrlFlag, cmd.PersistentFlags().Lookup(AwsSqsUrlFlag))
	if err != nil {
		panic(err)
	}
	cmd.PersistentFlags().String(AwsRegionFlag, "", "Aws region. Default is empty - means autodetect from IMDS. Can be set using AWS_REGION env variable")
	err = viper.BindPFlag(AwsRegionFlag, cmd.PersistentFlags().Lookup(AwsRegionFlag))
	if err != nil {
		panic(err)
	}
	cmd.PersistentFlags().Int(DrainTimeoutFlag, 180, "Timeout of node drain. Can be set using DRAIN_TIMEOUT env variable")
	err = viper.BindPFlag(DrainTimeoutFlag, cmd.PersistentFlags().Lookup(DrainTimeoutFlag))
	if err != nil {
		panic(err)
	}
	cmd.PersistentFlags().Int(PortFlag, 8080, "Http port (used for observability). Can be set using PORT env variable")
	err = viper.BindPFlag(PortFlag, cmd.PersistentFlags().Lookup(PortFlag))
	if err != nil {
		panic(err)
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
