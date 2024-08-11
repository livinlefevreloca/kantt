package storage

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Pod struct {
	gorm.Model
	Name         string `gorm:"uniqueIndex:pod_name_namespace_uniq;not null"`
	Namespace    string `gorm:"uniqueIndex:pod_name_namespace_uniq;not null"`
	OwnerID      int
	Owner        Owner
	PendingTime  time.Time `gorm:"default:null;index:pod_pending_time_idx"`
	StartingTime time.Time `gorm:"default:null;index:pod_starting_time_idx"`
	EndingTime   time.Time `gorm:"default:null;index:pod_ending_time_idx"`
}

// A pod is tied to a deployment, statefulset, or daemonset
// through its orchestrator record
type Owner struct {
	gorm.Model
	Name      string `gorm:"uniqueIndex:owner_name_namespace_uniq;not null"`
	Namespace string `gorm:"uniqueIndex:owner_name_namespace_uniq;not null"`
	Kind      string
}

// A pod is tied to a node for a specific period of time
// through a NodePod record
type NodePod struct {
	gorm.Model
	NodeID    int
	Node      Node
	PodID     int
	Pod       Pod
	StartTime time.Time `gorm:"default:null"`
	EndTime   time.Time `gorm:"default:null"`
}

type Node struct {
	gorm.Model
	Name string
	IP   string
}

// Add coalesce index
func AddIndexOnExpression(db *gorm.DB) {
	sql := `
    CREATE INDEX IF NOT EXISTS pod_times_coalesce_idx
    ON pods (COALESCE(pending_time, starting_time, ending_time))
    `
	result := db.Exec(sql)
	if result.Error != nil {
		fmt.Println("Error creating index:", result.Error)
	} else {
		fmt.Println("Index created successfully")
	}
}
