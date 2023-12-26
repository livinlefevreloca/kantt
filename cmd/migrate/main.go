package main

import (
	"github.com/endophage/kant/lib/config"
	"github.com/endophage/kant/lib/storage"
)

func main() {
	db := config.Database()
	// Migrate the schema
	db.AutoMigrate(storage.Resource{})
	db.AutoMigrate(storage.Event{})
}
