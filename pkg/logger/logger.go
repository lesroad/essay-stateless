package logger

import (
	"essay-stateless/internal/config"

	"github.com/sirupsen/logrus"
)

func Init(config config.LogConfig) {
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}

	logrus.SetLevel(level)

	if config.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}
