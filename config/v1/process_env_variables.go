package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Process - populates the specified struct based on environment variables
func Process(conf interface{}) {
	err := envconfig.Process("", conf)
	if err != nil {
		panic(err)
	}
}
