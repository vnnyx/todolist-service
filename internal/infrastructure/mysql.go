package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/sirupsen/logrus"
)

func NewMySQLDatabase(cfg *Config) *sql.DB {
	ctx, cancel := NewMySQLContext()
	defer cancel()

	mysqlHostSlave := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true",
		cfg.MysqlUser,
		cfg.MysqlPassword,
		cfg.MysqlHost,
		cfg.MysqlPort,
		cfg.MysqlDBName,
	)

	sqlDB, err := sql.Open("mysql", mysqlHostSlave)
	if err != nil {
		logrus.Fatal(err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	mysqlPoolMax := cfg.MysqlPoolMax

	mysqlIdleMax := cfg.MysqlIdleMax

	mysqlMaxLifeTime := cfg.MysqlMaxLifeTimeMinute

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(mysqlIdleMax)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(mysqlPoolMax)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(mysqlMaxLifeTime) * time.Minute)

	//sqlDB.SetConnMaxIdleTime(time.Duration(mysqlMaxIdleTime) * time.Minute)

	return sqlDB
}

func NewMySQLContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
