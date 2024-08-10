package main

import (
	"github.com/livinlefevreloca/kantt/pkg/config"
)

func main() {
	db := config.Database()
	config.RunMigrations(db)
}
