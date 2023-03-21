package infrastructure

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	AppPort                string `mapstructure:"APP_PORT"`
	MysqlPoolMin           int    `mapstructure:"MYSQL_POOL_MIN"`
	MysqlPoolMax           int    `mapstructure:"MYSQL_POOL_MAX"`
	MysqlIdleMax           int    `mapstructure:"MYSQL_IDLE_MAX"`
	MysqlMaxIdleTimeMinute int    `mapstructure:"MYSQL_MAX_IDLE_TIME_MINUTE"`
	MysqlMaxLifeTimeMinute int    `mapstructure:"MYSQL_MAX_LIFE_TIME_MINUTE"`
	RedisHost              string `mapstructure:"REDIS_HOST"`
	RedisPassword          string `mapstructure:"REDIS_PASSWORD"`
	MysqlHost              string `mapstructure:"MYSQL_HOST"`
	MysqlPort              int    `mapstructure:"MYSQL_PORT"`
	MysqlUser              string `mapstructure:"MYSQL_USER"`
	MysqlPassword          string `mapstructure:"MYSQL_PASSWORD"`
	MysqlDBName            string `mapstructure:"MYSQL_DBNAME"`
}

func NewConfig(configName string) *Config {
	config := &Config{}
	viper.AddConfigPath(".")
	viper.SetConfigName(configName)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		logrus.Fatal(err)
	}
	return config
}
