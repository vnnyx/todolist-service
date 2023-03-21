package bootstrap

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func RunMigration() {
	cfg := infrastructure.NewConfig(".env")
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true",
		cfg.MysqlUser,
		cfg.MysqlPassword,
		cfg.MysqlHost,
		cfg.MysqlPort,
		cfg.MysqlDBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Fatal(err)
	}

	db.AutoMigrate(&entity.Activity{}, &entity.Todo{})
}
