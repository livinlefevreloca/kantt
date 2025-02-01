package main

import (
	"database/sql"
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
	r.Use(CORSMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	handler := Handler{db: db}

	r.GET("/pods", handler.podsHandler)
	r.GET("/nodes", handler.NodesHandler)
	r.GET("/namespaces", handler.NamespacesHandler)

	r.Run()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

type PodInfo struct {
	Name       sql.NullString `json:"name"`
	Namespace  sql.NullString `json:"namespace"`
	CreateTime sql.NullTime   `json:"createtime"`
	DeleteTime sql.NullTime   `json:"deletetime"`
	OwnerName  sql.NullString `json:"ownername"`
	OwnerKind  sql.NullString `json:"ownerkind"`
	NodeName   sql.NullString `json:"nodename"`
	NodeIP     sql.NullString `json:"nodeip"`
}

type NodeInfo struct {
	Name sql.NullString `json:"name"`
}

type NamespaceInfo struct {
	Namespace sql.NullString `json:"namespace"`
}

type Handler struct {
	db *gorm.DB
}

func parseTime(timeStr string) (time.Time, error) {
	timeStamp, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(timeStamp, 0), nil
}

func (h *Handler) podsHandler(c *gin.Context) {
	db := h.db
	startTimeStr := c.Query("startTime")
	endTimeStr := c.Query("endTime")
	namespace := c.DefaultQuery("namespace", "all")
	node := c.DefaultQuery("node", "all")

	startTime, err := parseTime(startTimeStr)
	if err != nil {
		slog.Error("Error parsing startTimeStamp", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid startTime parameter",
		})
		return
	}

	endTime, err := parseTime(endTimeStr)
	if err != nil {
		slog.Error("Error parsing endTimeStamp", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid endTime parameter",
		})
		return
	}

	query := db.Model(&storage.Pod{}).
		Select(`
				pods.name as "Name"
				, pods.namespace as "Namespace"
				, pods.create_time as "CreateTime"
				, pods.delete_time as "DeleteTime"
				, owners.name as "OwnerName"
				, owners.kind as "OwnerKind"
				, nodes.name as "NodeName"
				, nodes.ip as "NodeIP"
			`).
		Joins("JOIN owners ON pods.owner_id = owners.id").
		Joins("JOIN node_pods ON pods.id = node_pods.pod_id").
		Joins("JOIN nodes ON node_pods.node_id = nodes.id").
		Where(`
				pods.create_time <= ? AND
				(pods.delete_time >= ? OR pods.delete_time IS NULL)
				`,
			endTime,
			startTime,
		)
	if namespace != "all" {
		query.Where("pods.namespace = ?", namespace)
	}

	if node != "all" {
		query.Where("nodes.name = ?", node)
	}

	var pods []PodInfo
	query.Find(&pods)
	c.JSON(http.StatusOK, gin.H{
		"pods": pods,
	})
}

func (h *Handler) NodesHandler(c *gin.Context) {
	db := h.db
	startTimeStr := c.Query("startTime")
	endTimeStr := c.Query("endTime")
	namespace := c.DefaultQuery("namespace", "all")

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

	query := db.Model(&storage.Node{}).
		Select(`distinct nodes.name as "Name"`).
		Joins("JOIN node_pods ON nodes.id = node_pods.node_id").
		Joins("JOIN pods ON node_pods.pod_id = pods.id").
		Where(`
			pods.create_time <= ? AND
			(pods.delete_time >= ? OR pods.delete_time IS NULL)
			`,
			startTime,
			endTime,
		)

	if namespace != "all" {
		query.Where("pods.namespace = ?", namespace)
	}

	var nodes []NodeInfo
	query.Find(&nodes)
	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
	})

}

func (h *Handler) NamespacesHandler(c *gin.Context) {
	db := h.db
	startTimeStr := c.Query("startTime")
	endTimeStr := c.Query("endTime")
	node := c.DefaultQuery("node", "all")

	startTime, err := parseTime(startTimeStr)
	if err != nil {
		slog.Error("Error parsing startTimeStamp", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid startTime parameter",
		})
		return
	}

	endTime, err := parseTime(endTimeStr)
	if err != nil {
		slog.Error("Error parsing endTimeStamp", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid endTime parameter",
		})
		return
	}

	query := db.Model(&storage.Pod{}).
		Select(`
				distinct pods.namespace as "Namespace"
			`)

	if node != "all" {
		query = query.Joins("JOIN node_pods ON pods.id = node_pods.pod_id").
			Joins("JOIN nodes ON node_pods.node_id = nodes.id").
			Where("nodes.name = ?", node)
	}

	query = query.Where(`
		pods.create_time <= ? AND
		(pods.delete_time >= ? OR pods.delete_time IS NULL)
		`,
		startTime,
		endTime,
	)

	var namespaces []NamespaceInfo
	query.Find(&namespaces)
	c.JSON(http.StatusOK, gin.H{
		"namespaces": namespaces,
	})
}
