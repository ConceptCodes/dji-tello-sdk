package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetFormatter(&logrus.TextFormatter{})
	logLevel := os.Getenv("TELLO_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	SetLogLevel(logLevel)
}

func SetLogLevel(level string) {
	switch level {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.Warn("Invalid log level, defaulting to INFO")
		Logger.SetLevel(logrus.InfoLevel)
	}
}
