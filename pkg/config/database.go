package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/livinlefevreloca/kantt/pkg/storage"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
	err := db.Debug().AutoMigrate(&storage.Pod{})
	if err != nil {
		slog.Error("Error migrating Pod", "Error", err)
		os.Exit(1)
	}
	db.Debug().AutoMigrate(&storage.Owner{})
	if err != nil {
		slog.Error("Error migrating Owner", "Error", err)
		os.Exit(1)
	}
	db.Debug().AutoMigrate(&storage.Node{})
	if err != nil {
		slog.Error("Error migrating Node", "Error", err)
		os.Exit(1)
	}
	db.Debug().AutoMigrate(&storage.NodePod{})
	if err != nil {
		slog.Error("Error migrating NodePod", "Error", err)
		os.Exit(1)
	}

	storage.AddIndexOnExpression(db)
}

func Database() *gorm.DB {
	var (
		engine = viper.GetString("database.engine")
		db     *gorm.DB
		err    error
	)
	fmt.Println("Database engine: ", engine)
	if engine == "postgres" {
		db, err = gorm.Open(
			postgres.Open(
				fmt.Sprintf("host=%s user=%s password=%s database=%s port=%s sslmode=%s TimeZone=%s",
					viper.GetString("database.host"),
					viper.GetString("database.user"),
					viper.GetString("database.password"),
					viper.GetString("database.database"),
					viper.GetString("database.port"),
					viper.GetString("database.sslmode"),
					viper.GetString("database.timezone"),
				),
			),
			&gorm.Config{},
		)
		RunMigrations(db)
	} else {
		slog.Error("Unsupported database engine", "Engine", engine)
		err = fmt.Errorf("unsupported database engine: %s", engine)
	}
	if err != nil {
		panic(fmt.Errorf("failed to connect database: %w", err))
	}
	return db
}
