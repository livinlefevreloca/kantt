package main

import (
	"github.com/endophage/kantt/pkg/config"
	"github.com/endophage/kantt/pkg/storage"
)

func main() {
	db := config.Database()
	// Migrate the schema
	db.AutoMigrate(storage.Resource{})
	db.AutoMigrate(storage.Event{})
}
