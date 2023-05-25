package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupFlags(t *testing.T) {
	cmd := &cobra.Command{}
	res := SetupFlags(cmd)
	assert.NoError(t, res)
}

func TestValdiateRootFlagsOk(t *testing.T) {
	viper.Set(LogLevelFlag, "info")
	viper.Set(LogFormatFlag, "json")
	viper.Set(NodeNameFlag, "test-node")
	res := ValidateRootFlags()

	assert.NoError(t, res)
}

func TestValdiateRootFlagsFail(t *testing.T) {
	viper.Set(LogLevelFlag, "wrong")
	res := ValidateRootFlags()

	assert.Error(t, res)
}
