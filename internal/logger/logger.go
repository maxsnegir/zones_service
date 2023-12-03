package logger

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/maxsnegir/zones_service/internal/config"
)

func New(env string) *logrus.Logger {
	switch env {
	case config.EnvLocal, config.EnvTest:
		return &logrus.Logger{
			Out:       os.Stdout,
			Formatter: &logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05"},
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.DebugLevel,
		}
	case config.EnvDev:
		return &logrus.Logger{
			Out:       os.Stdout,
			Formatter: &logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05"},
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		}
	case config.EnvProd:
		return &logrus.Logger{
			Out:       os.Stdout,
			Formatter: &logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05"},
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		}
	default:
		return &logrus.Logger{
			Out:       os.Stdout,
			Formatter: &logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05"},
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		}
	}
}
