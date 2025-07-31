package handlers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/BramseDev/imageAnalyzer/monitoring"
	"github.com/gin-gonic/gin"
)

type ActiveConnectionTracker struct {
	mu          sync.RWMutex
	connections map[string]time.Time
	metrics     *monitoring.Metrics
}

func NewActiveConnectionTracker(metrics *monitoring.Metrics) *ActiveConnectionTracker {
	tracker := &ActiveConnectionTracker{
		connections: make(map[string]time.Time),
		metrics:     metrics,
	}

	// Cleanup expired connections every 30 seconds
	go tracker.cleanupExpiredConnections()

	return tracker
}

func (t *ActiveConnectionTracker) TrackConnection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Erstelle bessere Session-ID
		userAgent := c.Request.UserAgent()
		clientIP := c.ClientIP()

		// Nur Dashboard-Zugriffe zählen als "aktive User"
		if strings.Contains(c.Request.URL.Path, "/dashboard") ||
			strings.Contains(c.Request.URL.Path, "/metrics") ||
			strings.Contains(c.Request.URL.Path, "/upload") {

			connID := clientIP + ":" + userAgent

			t.mu.Lock()
			t.connections[connID] = time.Now()
			activeCount := int64(len(t.connections))
			t.mu.Unlock()

			// Debug logging
			fmt.Printf("CONNECTION TRACKER: User activity detected, %d active users\n", activeCount)

			t.metrics.UpdateActiveConnections(activeCount)
		}

		c.Next()
	}
}

func (t *ActiveConnectionTracker) cleanupExpiredConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t.mu.Lock()
		cutoff := time.Now().Add(-2 * time.Minute) // ← Von 5 auf 2 Minuten reduziert

		for connID, lastSeen := range t.connections {
			if lastSeen.Before(cutoff) {
				delete(t.connections, connID)
			}
		}

		activeCount := int64(len(t.connections))
		t.mu.Unlock()

		t.metrics.UpdateActiveConnections(activeCount)

		// Debug logging
		fmt.Printf("CONNECTION TRACKER: %d active connections after cleanup\n", activeCount)
	}
}
