package main

import (
	"log"
	"log/slog"

	"github.com/BramseDev/imageAnalyzer/dashboard"
	"github.com/BramseDev/imageAnalyzer/internal/handlers"
	"github.com/BramseDev/imageAnalyzer/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	customLogger := logging.NewLogger(slog.LevelInfo)
	r := gin.Default()

	r.LoadHTMLGlob("dashboard/templates/*")

	// Handler registrieren und Metrics holen
	metrics := handlers.RegisterHandlers(r, customLogger)

	// Connection Tracker als globale Middleware (vor allen Routes)
	connectionTracker := handlers.NewActiveConnectionTracker(metrics)
	r.Use(connectionTracker.TrackConnection())

	// Dashboard Routes registrieren
	dashboard.RegisterDashboardRoutes(r)

	customLogger.Info("Server starting", "port", 8080)
	log.Println("Metrics Dashboard: http://localhost:8080/dashboard/metrics")
	log.Println("Health Dashboard: http://localhost:8080/dashboard/health")

	r.Run(":8080")
}
