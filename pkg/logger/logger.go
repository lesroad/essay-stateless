package logger

import (
	"essay-stateless/internal/config"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

func Init(config config.LogConfig) {
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}

	logrus.SetLevel(level)

	logrus.SetReportCaller(true)

	if config.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return f.Function, fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return f.Function, fmt.Sprintf("%s:%d", filename, f.Line)
			},
			FullTimestamp: true,
		})
	}
}
