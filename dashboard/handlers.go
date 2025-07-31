package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Registriere Dashboard-Routes
func RegisterDashboardRoutes(r *gin.Engine) {
	// Statische Dateien servieren
	r.Static("/dashboard/static", "./dashboard/static")

	// Dashboard-Routes
	r.GET("/dashboard/metrics", metricsPageHandler)
	r.GET("/dashboard/health", healthPageHandler)
}

func metricsPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "metrics.html", gin.H{
		"title": "Metrics Dashboard",
	})
}

func healthPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "health.html", gin.H{
		"title": "Health Dashboard",
	})
}
