package database

type ConnectionConfig struct {
	PostgresPort     string `default:"5432" envconfig:"POSTGRES_PORT"`
	PostgresHostname string `default:"localhost" envconfig:"POSTGRES_HOSTNAME"`
	PostgresUser     string `default:"postgres" envconfig:"POSTGRES_USER"`
	PostgresPassword string `default:"postgres" envconfig:"POSTGRES_PASSWORD"`
	PostgresDBName   string `default:"postgres" envconfig:"POSTGRES_DB"`
	PostgresSSLMode  string `default:"disable" envconfig:"POSTGRES_SSL_MODE"`
	PostgresSchema   string `default:"public" envconfig:"POSTGRES_SCHEMA"`
}

type Config struct {
	MaxOpenConns string `default:"10" envconfig:"PG_MAX_OPEN_CONNS"`
	MaxIdleConns string `default:"2" envconfig:"PG_MAX_IDLE_CONNS"`
}
