package bootstrap

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
)

func RunMigration() {
	cfg := infrastructure.NewConfig(".env")
	migration, err := migrate.New(cfg.MigrationSource, fmt.Sprintf("mysql://%v:%v@tcp(%v:%v)/%v?parseTime=true",
		cfg.MysqlUser,
		cfg.MysqlPassword,
		cfg.MysqlHost,
		cfg.MysqlPort,
		cfg.MysqlDBName,
	))
	if err != nil {
		logrus.Fatal(err)
	}
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		logrus.Fatal(err)
	}
}
