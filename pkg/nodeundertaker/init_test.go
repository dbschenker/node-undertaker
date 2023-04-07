package nodeundertaker

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	viper.Reset()
	logrus.Infof("Initialized tests")
}
