package storage

import "gorm.io/gorm"

type Resource struct {
	gorm.Model
	Name      string
	Kind      string
	Namespace string
}

type Event struct {
	gorm.Model
	Type     string
	Resource Resource
}
