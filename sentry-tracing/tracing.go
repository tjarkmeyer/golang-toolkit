package tracing

import (
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

func InitSentry(dsn, env, serverName string) (*sentryhttp.Handler, error) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		ServerName:       serverName,
		TracesSampleRate: 1.0,
	})

	if err != nil {
		return nil, err
	}

	return sentryhttp.New(sentryhttp.Options{}), nil
}
