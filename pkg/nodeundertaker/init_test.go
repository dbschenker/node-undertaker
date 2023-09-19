package nodeundertaker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	viper.Reset()
	prometheus.NewRegistry()
	logrus.Infof("Initialized tests")
}
