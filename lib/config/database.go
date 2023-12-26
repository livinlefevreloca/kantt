package config

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Database() *gorm.DB {
	var (
		engine = viper.GetString("database.engine")
		db     *gorm.DB
		err    error
	)
	if engine == "sqlite" {
		db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	} else if engine == "postgres" {
		db, err = gorm.Open(
			postgres.Open(
				fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
					viper.GetString("database.host"),
					viper.GetString("database.user"),
					viper.GetString("database.password"),
					viper.GetString("database.dbname"),
					viper.GetString("database.port"),
					viper.GetString("database.sslmode"),
					viper.GetString("database.timezone"),
				),
			),
			&gorm.Config{},
		)
	}
	if err != nil {
		panic(fmt.Errorf("failed to connect database: %w", err))
	}
	return db
}
