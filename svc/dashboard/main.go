package main

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livinlefevreloca/kantt/pkg/config"
	"github.com/livinlefevreloca/kantt/pkg/storage"
	"gorm.io/gorm"
)

func main() {
	db := config.Database()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	handler := Handler{db: db}

	r.GET("/pods", handler.podsHandler)

	r.Run()
}

type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	OwnerName string `json:"ownerName"`
	OwnerKind string `json:"ownerKind"`
	NodeName  string `json:"nodeName"`
	NodeIP    string `json:"nodeIP"`
}

type Handler struct {
	db *gorm.DB
}

func (h *Handler) podsHandler(c *gin.Context) {
	db := h.db
	startTimeStr := c.Query("startTime")
	endTimeStr := c.Query("endTime")
	namespace := c.DefaultQuery("namespace", "all")
	node := c.DefaultQuery("node", "all")
	owner := c.DefaultQuery("owner", "all")

	startTimeStamp, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		slog.Error("Error parsing startTimeStamp", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid startTime parameter",
		})
		return
	}
	startTime := time.Unix(startTimeStamp, 0)

	endTimeStamp, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		slog.Error("Error parsing endTimeStamp", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid endTime parameter",
		})
		return
	}
	endTime := time.Unix(endTimeStamp, 0)

	query := db.Model(&storage.Pod{}).
		Select(`
				pods.name
				, pods.namespace
				, owners.name as owner_name
				, owners.kind as owner_kind
				, nodes.name as node_name
				, nodes.ip as node_ip
			`).
		Joins("JOIN owners ON pods.owner_id = owners.id").
		Joins("JOIN node_pods ON pods.id = node_pods.pod_id").
		Joins("JOIN nodes ON node_pods.node_id = nodes.id").
		Where(`
				coalesce(pending_time, starting_time, ending_time) >= ?
				AND coalesce(pending_time, starting_time, ending_time) <= ?
				`,
			startTime,
			endTime,
		)
	var pods []PodInfo
	if namespace != "all" {
		query.Where("pods.namespace = ?", namespace)
	} else if node != "all" {
		query.Where("nodes.name = ?", node)
	} else if owner != "all" {
		query.Where("owners.name = ?", owner)
	}

	query.Find(&pods)
	c.JSON(http.StatusOK, gin.H{
		"pods": pods,
	})
}
