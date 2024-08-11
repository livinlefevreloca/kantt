package config

import (
	"fmt"

	"github.com/livinlefevreloca/kantt/pkg/storage"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
	db.AutoMigrate(&storage.Pod{})
	db.AutoMigrate(&storage.Owner{})
	db.AutoMigrate(&storage.Node{})
	db.AutoMigrate(&storage.NodePod{})

	storage.AddIndexOnExpression(db)
}

func Database() *gorm.DB {
	var (
		engine = viper.GetString("database.engine")
		db     *gorm.DB
		err    error
	)
	fmt.Println("engine: ", engine)
	if engine == "sqlite" {
		db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
		// If running with sqlite we need to migrate the schema on startup
		RunMigrations(db)
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
