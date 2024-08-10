package storage

import (
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

type Resource struct {
	gorm.Model
	Name string
	schema.GroupVersionKind
	Namespace string
}

type Event struct {
	gorm.Model
	Type       watch.EventType
	ResourceID int
	Resource   Resource
}
