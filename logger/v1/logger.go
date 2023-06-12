package logger

import (
	"encoding/json"

	"github.com/tjarkmeyer/golang-toolkit/config/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Level            string      `json:"level" default:"debug" envconfig:"LOG_LEVEL"`
	Encoding         string      `json:"encoding" default:"json" envconfig:"LOG_ENCODING"`
	OutputPaths      []string    `json:"outputPaths" default:"stdout" envconfig:"LOG_OUTPUT_PATHS"`
	ErrorOutputPaths []string    `json:"errorOutputPaths" default:"stderr" envconfig:"LOG_ERROR_OUTPUT_PATHS"`
	InitialFields    interface{} `json:"initialFields" envconfig:"LOG_INTIAL_FIELDS"`
	EncoderConfig    struct {
		MessageKey   string `json:"messageKey" default:"message" envconfig:"LOG_MESSAGE_KEY"`
		LevelKey     string `json:"levelKey" default:"level" envconfig:"LOG_LEVEL_KEY"`
		LevelEncoder string `json:"levelEncoder" default:"lowercase" envconfig:"LOG_LEVEL_ENCODER"`
		TimeKey      string `json:"timeKey" default:"timestamp" envconfig:"LOG_TIME_KEY"`
	}
}

func NewLogger(appVersion string) *zap.Logger {
	var ownZapConfig LogConfig

	config.Process(&ownZapConfig)
	zap := convertOwnConfigToZapConfig(ownZapConfig)
	zap.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zap.InitialFields = map[string]interface{}{
		"version": appVersion,
	}

	logger, err := zap.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

func convertOwnConfigToZapConfig(config LogConfig) zap.Config {
	rawConfig, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	var zapConfig zap.Config
	if err := json.Unmarshal(rawConfig, &zapConfig); err != nil {
		panic(err)
	}

	return zapConfig
}
