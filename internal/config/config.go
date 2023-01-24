package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Settings struct {
	Debug          bool   `default:"false"`
	PostgresDSN    string `split_words:"true" default:"postgres://postgres:root@postgres:5432/cyberloop-edr?sslmode=disable"`
	ResendInterval uint   `split_words:"true" default:"20"`
	Addr           string `default:"grpc_server:50051"`
}

var settings Settings

func init() {
	if err := envconfig.Process("edr", &settings); err != nil {
		logrus.Fatalln("cannot apply configuration:", err)
	}

	logrus.Infof("Using configuration %+v", settings)

	if settings.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func Get() *Settings {
	return &settings
}

func (s *Settings) GetResendInterval() time.Duration {
	return time.Duration(s.ResendInterval) * time.Second
}
