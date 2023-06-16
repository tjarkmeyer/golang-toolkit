package database

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dsnTemplate = "host=%s port=%s user=%s dbname=%s password=%s sslmode=%s search_path=%s"

func Connect(db ConnectionConfig, c Config) *gorm.DB {
	dsn := fmt.Sprintf(dsnTemplate, db.PostgresHostname, db.PostgresPort, db.PostgresUser, db.PostgresDBName, db.PostgresPassword, db.PostgresSSLMode, db.PostgresSchema)
	pgDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	sqlDb, err := pgDB.DB()
	if err != nil {
		panic(err)
	}

	maxIdleCons, err := strconv.Atoi(c.MaxIdleConns)
	if err != nil {
		panic(err)
	}

	maxOpenCons, err := strconv.Atoi(c.MaxOpenConns)
	if err != nil {
		panic(err)
	}

	sqlDb.SetMaxIdleConns(maxIdleCons)
	sqlDb.SetMaxOpenConns(maxOpenCons)
	sqlDb.SetConnMaxLifetime(time.Hour)

	return pgDB
}
