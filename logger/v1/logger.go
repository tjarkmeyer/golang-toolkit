package logger

import (
	"os"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(env, dsn string) *zap.Logger {
	encoder := zapEncoder(env)
	logLevel := zapLogLevel(env)
	logWriter := zapLogWriter()
	core := zapcore.NewCore(encoder, logWriter, logLevel)
	logger := zap.New(core, zap.AddCaller())
	return attachSentryLogger(logger, newSentryClientFromDSN(dsn, env))
}

func zapEncoder(env string) zapcore.Encoder {
	var encoderConfig zapcore.EncoderConfig

	if env == "production" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func zapLogWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}

func zapLogLevel(env string) zapcore.LevelEnabler {
	if env == "production" {
		return zapcore.ErrorLevel
	} else {
		return zapcore.DebugLevel
	}
}

func attachSentryLogger(log *zap.Logger, clientFactory zapsentry.SentryClientFactory) *zap.Logger {
	cfg := zapsentry.Configuration{
		Level: zapcore.ErrorLevel,
	}
	core, err := zapsentry.NewCore(cfg, clientFactory)

	if err != nil {
		panic(err)
	}

	return zapsentry.AttachCoreToLogger(core, log)
}

func newSentryClientFromDSN(DSN, env string) zapsentry.SentryClientFactory {
	return func() (*sentry.Client, error) {
		return sentry.NewClient(sentry.ClientOptions{
			Dsn:         DSN,
			Environment: env,
		})
	}
}
