package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/analysis"
	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
	"github.com/BramseDev/imageAnalyzer/internal/handlers/verdict"
	"github.com/BramseDev/imageAnalyzer/logging"
	"github.com/BramseDev/imageAnalyzer/monitoring"
	"github.com/gin-gonic/gin"
)

const (
	MaxFiles          = 3
	MaxFileSize       = 50 * 1024 * 1024
	MaxFilenameLength = 255
)

var (
	uploadLimiter = make(chan struct{}, MaxFiles)
	metrics       *monitoring.Metrics
	globalLogger  *logging.Logger
)

func init() {
	metrics = monitoring.NewMetrics()
}

func SetGlobalLogger(logger *logging.Logger) {
	globalLogger = logger
}

func uploadHandler(c *gin.Context) {
	startTime := time.Now()

	if metrics != nil {
		metrics.IncrementActiveConnections()
		defer metrics.DecrementActiveConnections()
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		if metrics != nil {
			metrics.RecordError("upload")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	if err := utils.ValidateFile(header); err != nil {
		if metrics != nil {
			metrics.RecordError("upload")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tempFilePath, err := utils.CreateSecureTempFile(file, header)
	if err != nil {
		if metrics != nil {
			metrics.RecordError("upload")
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer os.Remove(tempFilePath)

	if err := utils.ValidateFileContent(tempFilePath); err != nil {
		if metrics != nil {
			metrics.RecordError("upload")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pipelineStart := time.Now()
	results, err := analysis.RunSecureAnalyses(tempFilePath, globalLogger.Logger)
	pipelineDuration := time.Since(pipelineStart)

	if err != nil {
		if metrics != nil {
			metrics.RecordError("pipeline")
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed"})
		return
	}

	if metrics != nil {
		metrics.RecordSuccess("upload")
		metrics.RecordDuration("upload", time.Since(startTime))
		metrics.RecordSuccess("pipeline")
		metrics.RecordDuration("pipeline", pipelineDuration)

		if results.CacheHit {
			metrics.RecordCacheHit()
		} else {
			metrics.RecordCacheMiss()
		}
	}

	verdictData := verdict.CalculateOverallVerdict(results)

	if verdictString, exists := verdictData["verdict"].(string); exists {
		fmt.Printf("DEBUG uploadHandler: Calling RecordVerdict with: %s\n", verdictString)
		if metrics != nil {
			metrics.RecordVerdict(verdictString, results.EarlyExit)
		} else {
			fmt.Printf("DEBUG: metrics is nil!\n")
		}
	} else {
		fmt.Printf("DEBUG: No verdict string found in: %+v\n", verdictData)
	}

	response := analysis.CreateStructuredResponse(results)
	response["analysis"] = verdictData
	response["metadata"] = gin.H{
		"analysis_duration": time.Since(startTime).Milliseconds(),
		"pipeline_duration": pipelineDuration.Milliseconds(),
		"early_exit":        results.EarlyExit,
		"cache_hit":         results.CacheHit,
		"analyses_run":      len(results.Results),
	}

	c.JSON(http.StatusOK, response)
}

func metricsHandler(c *gin.Context) {
	fmt.Printf("DEBUG: metrics variable: %+v\n", metrics)

	if metrics == nil {
		fmt.Printf("DEBUG: metrics is nil!\n")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "metrics not initialized"})
		return
	}

	response := metrics.GetMetricsSummary()
	fmt.Printf("DEBUG: Full response: %+v\n", response)

	response["timestamp"] = time.Now().Unix()

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}

func healthHandler(c *gin.Context) {
	summary := metrics.GetSummary()
	healthy := true
	issues := []string{}

	if overall, ok := summary["overall"].(map[string]interface{}); ok {
		if errorRate, ok := overall["overall_error_rate"].(float64); ok && errorRate > 0.1 {
			healthy = false
			issues = append(issues, "High error rate detected")
		}
	}

	avgResponseTime := int64(0)
	metricsSummary := metrics.GetMetricsSummary()
	if upload, exists := metricsSummary["upload"].(map[string]interface{}); exists {
		if avgDuration, exists := upload["average_duration"].(int64); exists {
			avgResponseTime = avgDuration
		}
	}

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"healthy":             healthy,
		"issues":              issues,
		"metrics":             summary,
		"average_response_ms": avgResponseTime,
		"timestamp":           time.Now().Unix(),
	})
}

func RegisterHandlers(r *gin.Engine, logger *logging.Logger) *monitoring.Metrics {
	SetGlobalLogger(logger)
	r.POST("/upload", uploadHandler)
	r.GET("/metrics", metricsHandler)
	r.GET("/health", healthHandler)
	return metrics
}

// package handlers

// import (
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"math"
// 	"mime/multipart"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"time"

// 	"github.com/BramseDev/imageAnalyzer/logging"
// 	"github.com/BramseDev/imageAnalyzer/monitoring"
// 	analyzer "github.com/BramseDev/imageAnalyzer/pkg/analyzer/pipeline"
// 	"github.com/BramseDev/imageAnalyzer/pkg/pythonrunner"
// 	"github.com/gin-gonic/gin"
// )

// const (
// 	MaxFileSize = 50 << 20 // 50MB
// 	MaxFiles    = 10       // Max concurrent uploads
// )

// var (
// 	uploadLimiter = make(chan struct{}, MaxFiles)
// 	metrics       *monitoring.Metrics
// )

// func init() {
// 	metrics = monitoring.NewMetrics()
// }

// var globalLogger *logging.Logger

// func SetGlobalLogger(logger *logging.Logger) {
// 	globalLogger = logger
// }

// func uploadHandler(c *gin.Context) {
// 	requestID := c.GetHeader("X-Request-ID")
// 	if requestID == "" {
// 		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
// 	}

// 	// Fallback falls globalLogger nil ist
// 	var logger *slog.Logger
// 	if globalLogger != nil {
// 		logger = globalLogger.With("request_id", requestID)
// 	} else {
// 		logger = slog.With("request_id", requestID)
// 	}

// 	requestStart := time.Now()
// 	logger.Info("Upload request started")
// 	// Rate limiting
// 	select {
// 	case uploadLimiter <- struct{}{}:
// 		defer func() { <-uploadLimiter }()
// 	case <-time.After(5 * time.Second):
// 		logger.Warn("Rate limit exceeded")
// 		metrics.RecordAnalysis("upload", time.Since(requestStart), fmt.Errorf("rate limit exceeded"))
// 		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Maximum number of concurrent uploads reached. Please try again later."})
// 		return
// 	}

// 	file, err := c.FormFile("image")
// 	if err != nil {
// 		logger.Error("No image file provided", "error", err)
// 		metrics.RecordAnalysis("upload", time.Since(requestStart), err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided. Please upload a valid image file."})
// 		return
// 	}

// 	logger.Info("File received", "filename", file.Filename, "size", file.Size)

// 	// Datei validieren
// 	if err := validateFile(file); err != nil {
// 		logger.Error("File validation failed", "error", err, "filename", file.Filename)
// 		metrics.RecordAnalysis("upload", time.Since(requestStart), err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Sichere temporäre Datei erstellen
// 	tempFile, cleanup, err := createSecureTempFile(file)
// 	if err != nil {
// 		logger.Error("Failed to create temp file", "error", err)
// 		metrics.RecordAnalysis("upload", time.Since(requestStart), err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to process uploaded file. Please try again."})
// 		return
// 	}
// 	defer cleanup()

// 	logger.Info("Temp file created", "path", tempFile)

// 	// Content-Type nochmals prüfen
// 	if err := validateFileContent(tempFile); err != nil {
// 		logger.Error("File content validation failed", "error", err)
// 		metrics.RecordAnalysis("upload", time.Since(requestStart), err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Analysen ausführen
// 	logger.Info("Starting analysis pipeline")
// 	analysisStart := time.Now()

// 	results := runSecureAnalyses(tempFile, logger)

// 	analysisTime := time.Since(analysisStart)
// 	logger.Info("Analysis completed", "duration", analysisTime)

// 	// Strukturierte Response mit User Explanations
// 	structuredResponse := createStructuredResponse(results)

// 	// Extract verdict for metrics
// 	verdict := calculateOverallVerdict(results)
// 	if verdictMap, ok := verdict["verdict"].(string); ok {
// 		// Record the verdict for AI detection metrics
// 		isEarlyExit := false
// 		if quality, exists := verdict["analysis_quality"].(string); exists {
// 			isEarlyExit = quality == "early_exit"
// 		}
// 		metrics.RecordVerdict(verdictMap, isEarlyExit)
// 	}

// 	// Gesamtzeit loggen
// 	totalTime := time.Since(requestStart)
// 	logger.Info("Request completed", "total_duration", totalTime, "analysis_duration", analysisTime)

// 	// Erfolgreiche Analyse in Metrics
// 	// metrics.RecordAnalysis("upload", totalTime, nil)

// 	c.JSON(http.StatusOK, structuredResponse)
// }

// func validateFile(file *multipart.FileHeader) error {
// 	if file.Size > MaxFileSize {
// 		return fmt.Errorf("file size exceeds maximum limit of %d MB. Please upload a smaller image", MaxFileSize/(1<<20))
// 	}

// 	if file.Size == 0 {
// 		return fmt.Errorf("uploaded file is empty. Please select a valid image file")
// 	}

// 	ext := strings.ToLower(filepath.Ext(file.Filename))
// 	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

// 	for _, validExt := range validExtensions {
// 		if ext == validExt {
// 			return nil
// 		}
// 	}

// 	return fmt.Errorf("file type '%s' is not supported. Allowed formats: JPG, JPEG, PNG, GIF, BMP, WEBP", ext)
// }

// func createSecureTempFile(file *multipart.FileHeader) (string, func(), error) {
// 	src, err := file.Open()
// 	if err != nil {
// 		return "", nil, fmt.Errorf("failed to open uploaded file: %w", err)
// 	}
// 	defer src.Close()

// 	tempFile, err := os.CreateTemp("", "analyzer_*."+strings.TrimPrefix(filepath.Ext(file.Filename), "."))
// 	if err != nil {
// 		return "", nil, fmt.Errorf("failed to create temporary file: %w", err)
// 	}

// 	if _, err := tempFile.ReadFrom(src); err != nil {
// 		tempFile.Close()
// 		os.Remove(tempFile.Name())
// 		return "", nil, fmt.Errorf("failed to save uploaded file: %w", err)
// 	}

// 	tempFile.Close()

// 	cleanup := func() {
// 		os.Remove(tempFile.Name())
// 	}

// 	return tempFile.Name(), cleanup, nil
// }

// func validateFileContent(filePath string) error {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return fmt.Errorf("unable to read uploaded file for validation")
// 	}
// 	defer file.Close()

// 	buffer := make([]byte, 512)
// 	_, err = file.Read(buffer)
// 	if err != nil {
// 		return fmt.Errorf("unable to analyze file content")
// 	}

// 	contentType := http.DetectContentType(buffer)
// 	validTypes := []string{
// 		"image/jpeg", "image/png", "image/gif",
// 		"image/webp", "image/bmp",
// 	}

// 	for _, validType := range validTypes {
// 		if contentType == validType {
// 			return nil
// 		}
// 	}

// 	return fmt.Errorf("file content does not match expected image format. Detected type: %s", contentType)
// }

// func runSecureAnalyses(filePath string, logger *slog.Logger) []pythonrunner.ScriptResult {
// 	analysisStart := time.Now()

// 	// Sicherheitschecks...
// 	absPath, err := filepath.Abs(filePath)
// 	if err != nil {
// 		metrics.RecordAnalysis("pipeline", time.Since(analysisStart), err)
// 		logger.Error("Invalid file path", "error", err)
// 		return []pythonrunner.ScriptResult{{Err: "Invalid file path provided"}}
// 	}

// 	tempDir := os.TempDir()
// 	if !strings.HasPrefix(absPath, tempDir) {
// 		securityErr := fmt.Errorf("file path security validation failed")
// 		metrics.RecordAnalysis("pipeline", time.Since(analysisStart), securityErr)
// 		logger.Error("File outside temp directory", "path", absPath, "temp_dir", tempDir)
// 		return []pythonrunner.ScriptResult{{Err: "File path security validation failed"}}
// 	}

// 	// Pipeline mit Cache erstellen
// 	pipeline := analyzer.NewAnalysisPipelineWithCache(metrics)
// 	pipeline.SetEarlyExitEnabled(false)

// 	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
// 	defer cancel()

// 	logger.Info("Starting pipeline analysis")
// 	pipelineResult, err := pipeline.RunAnalysis(ctx, absPath)

// 	// Metrics aufzeichnen
// 	metrics.RecordAnalysis("pipeline", time.Since(analysisStart), err)

// 	if err != nil {
// 		logger.Error("Pipeline failed", "error", err, "duration", time.Since(analysisStart))
// 		return []pythonrunner.ScriptResult{{Err: fmt.Sprintf("Analysis pipeline failed: %v", err)}}
// 	}

// 	// NEU: Detailliertes Logging
// 	logger.Info("Pipeline completed successfully",
// 		"duration", pipelineResult.ProcessTime,
// 		"stages_run", len(pipelineResult.StagesRun),
// 		"confidence", pipelineResult.Confidence,
// 		"cache_hit", pipelineResult.CacheHit)

// 	// NEU: Debug-Logging für verfügbare Stages
// 	fmt.Printf("\n=== PIPELINE DEBUG ===\n")
// 	fmt.Printf("Stages executed: %d\n", len(pipelineResult.StagesRun))
// 	fmt.Printf("Results available: %d\n", len(pipelineResult.Results))

// 	fmt.Printf("\nStages run:\n")
// 	for _, stage := range pipelineResult.StagesRun {
// 		fmt.Printf("  - %s\n", stage)
// 	}

// 	fmt.Printf("\nResults received:\n")
// 	for stageName, stageData := range pipelineResult.Results {
// 		fmt.Printf("  - %s: %T\n", stageName, stageData)

// 		// Debug: Prüfe ob Stage-Daten vorhanden sind
// 		if stageData == nil {
// 			fmt.Printf("    WARNING: %s has nil data\n", stageName)
// 		}
// 	}
// 	fmt.Printf("======================\n\n")

// 	// Update business metrics with pipeline result
// 	metrics.UpdateBusinessMetrics(pipelineResult)

// 	// Konvertiere Pipeline-Ergebnisse zu ScriptResult-Format
// 	var results []pythonrunner.ScriptResult
// 	for stageName, stageData := range pipelineResult.Results {
// 		results = append(results, pythonrunner.ScriptResult{
// 			Name: stageName,
// 			Data: stageData,
// 			Err:  "",
// 		})
// 	}

// 	// NEU: Debug-Logging für konvertierte Results
// 	fmt.Printf("\n=== CONVERTED RESULTS ===\n")
// 	for _, result := range results {
// 		fmt.Printf("  - %s: err='%s'\n", result.Name, result.Err)
// 	}
// 	fmt.Printf("=========================\n\n")

// 	return results
// }

// func createStructuredResponse(results []pythonrunner.ScriptResult) map[string]interface{} {
// 	analyses := make(map[string]interface{})
// 	for _, result := range results {
// 		if result.Err == "" {
// 			analyses[result.Name] = createAnalysisSection(result.Name, result.Data)
// 		}
// 	}

// 	verdict := calculateOverallVerdict(results)

// 	return map[string]interface{}{
// 		"analyses":        analyses,
// 		"overall_verdict": verdict,
// 		"timestamp":       time.Now().Unix(),
// 	}
// }

// func createAnalysisSection(analysisType string, data interface{}) map[string]interface{} {
// 	section := map[string]interface{}{
// 		"data":   data,
// 		"status": "success",
// 	}

// 	switch analysisType {
// 	case "metadata":
// 		section["display_name"] = "Image Metadata Analysis"
// 		section["description"] = "Technical image information and file properties"
// 		if dataMap, ok := data.(map[string]interface{}); ok {
// 			section["user_explanation"] = "Metadata analysis completed. This technical metadata provides evidence of image authenticity."
// 			if len(dataMap) > 5 {
// 				section["user_explanation"] = "Comprehensive metadata found containing technical information about image creation and file properties."
// 			} else if len(dataMap) > 0 {
// 				section["user_explanation"] = "Basic metadata present with limited technical information."
// 			}
// 		} else {
// 			section["user_explanation"] = "Metadata analysis completed."
// 		}

// 	case "exif":
// 		section["display_name"] = "EXIF Data Analysis"
// 		section["description"] = "Camera and capture settings information"
// 		section["user_explanation"] = generateEXIFExplanation(data)

// 	case "c2pa":
// 		section["display_name"] = "Content Provenance (C2PA)"
// 		section["description"] = "Cryptographic proof of content authenticity"
// 		section["user_explanation"] = generateC2PAExplanation(data)

// 	case "compression":
// 		section["display_name"] = "Compression Analysis"
// 		section["description"] = "File size and compression behavior analysis"
// 		section["user_explanation"] = generateCompressionExplanation(data)

// 	case "artifacts":
// 		section["display_name"] = "Visual Artifacts Detection"
// 		section["description"] = "Analysis of typical AI generation artifacts"
// 		section["user_explanation"] = generateArtifactsExplanation(data)

// 	case "pixel-analysis":
// 		section["display_name"] = "Pixel-Level Analysis"
// 		section["description"] = "Detailed pixel structure and mathematical properties"
// 		section["user_explanation"] = generatePixelExplanation(data)

// 	case "color-balance":
// 		section["display_name"] = "Color Distribution Analysis"
// 		section["description"] = "Color palette and balance characteristics"
// 		section["user_explanation"] = generateColorExplanation(data)

// 	case "advanced-artifacts":
// 		section["display_name"] = "Advanced AI Pattern Detection"
// 		section["description"] = "Specialized algorithms for subtle AI artifacts"
// 		section["user_explanation"] = generateAdvancedExplanation(data)

// 	case "object-coherence":
// 		section["display_name"] = "Object Consistency Analysis"
// 		section["description"] = "Logical plausibility of object sizes and arrangements"
// 		section["user_explanation"] = generateObjectCoherenceExplanation(data)

// 	case "lighting-analysis":
// 		section["display_name"] = "Lighting Consistency Analysis"
// 		section["description"] = "Physical consistency of light sources and shadows"
// 		section["user_explanation"] = generateLightingExplanation(data)

// 	case "ai-model":
// 		section["display_name"] = "Deep Learning AI Detection"
// 		section["description"] = "Neural network trained specifically on AI vs. real images"
// 		section["user_explanation"] = generateAIModelExplanation(data)

// 	default:
// 		section["display_name"] = strings.Title(analysisType)
// 		section["description"] = "Image analysis performed"
// 		section["user_explanation"] = "Analysis completed successfully"
// 	}

// 	return section
// }

// func generateEXIFExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if len(dataMap) == 0 {
// 			return "No EXIF data present. This is common for screenshots, heavily edited images, or AI-generated content."
// 		}

// 		cameraFields := []string{"Make", "Model", "DateTime", "Software", "GPS"}
// 		foundFields := 0

// 		for _, field := range cameraFields {
// 			if _, exists := dataMap[field]; exists {
// 				foundFields++
// 			}
// 		}

// 		if foundFields >= 3 {
// 			return fmt.Sprintf("Comprehensive EXIF data found (%d key fields). This typically indicates genuine camera hardware.", foundFields)
// 		}
// 	}

// 	return "EXIF data analysis completed. This technical metadata provides evidence of image authenticity."
// }

// func generateC2PAExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if score, exists := dataMap["score"]; exists {
// 			if scoreFloat, ok := score.(float64); ok {
// 				if scoreFloat >= 95 {
// 					return "C2PA certificate confirms AI generation with 100% certainty using cryptographic proof."
// 				}
// 				return fmt.Sprintf("C2PA analysis performed (Score: %.0f). Cryptographic signatures analyzed for provenance verification.", scoreFloat)
// 			}
// 		}
// 	}

// 	return "C2PA analysis completed. This system uses cryptographic proof to verify image origin."
// }

// func generateCompressionExplanation(data interface{}) string {
// 	return "Compression analysis performed. This measures information density and compressibility patterns."
// }

// func generateArtifactsExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if overall, exists := dataMap["overall_assessment"]; exists {
// 			if assessment, ok := overall.(map[string]interface{}); ok {
// 				if positiveCount, exists := assessment["positive_indicators"]; exists {
// 					if totalCount, exists := assessment["total_indicators"]; exists {
// 						if posFloat, ok := positiveCount.(float64); ok {
// 							if totFloat, ok := totalCount.(float64); ok {
// 								percentage := (posFloat / totFloat) * 100
// 								return fmt.Sprintf("Visual artifact analysis: %.0f of %.0f tests positive (%.0f%% AI probability). Examines ringing effects, color bleeding, and compression anomalies typical of neural networks.", posFloat, totFloat, percentage)
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return "Visual artifact analysis completed. Examines typical anomalies from AI generation processes."
// }

// func generatePixelExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if overall, exists := dataMap["overall_assessment"]; exists {
// 			if assessment, ok := overall.(map[string]interface{}); ok {
// 				if positiveCount, exists := assessment["positive_indicators"]; exists {
// 					if totalCount, exists := assessment["total_indicators"]; exists {
// 						if posFloat, ok := positiveCount.(float64); ok {
// 							if totFloat, ok := totalCount.(float64); ok {
// 								return fmt.Sprintf("Pixel analysis: %.0f of %.0f tests positive. Performs frequency analysis and noise characterization.", posFloat, totFloat)
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return "Pixel-level analysis completed. Examines mathematical properties of image structure."
// }

// func generateColorExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if aiScore, exists := dataMap["ai_color_score"]; exists {
// 			if score, ok := aiScore.(float64); ok {
// 				if score == 0 {
// 					return "Color analysis shows natural color distribution."
// 				} else if score < 0.4 {
// 					return fmt.Sprintf("Color analysis shows mostly natural distribution (%.0f%% AI probability).", score*100)
// 				} else {
// 					return fmt.Sprintf("Color analysis reveals unnatural color patterns (%.0f%% AI probability).", score*100)
// 				}
// 			}
// 		}
// 	}
// 	return "Color analysis completed. Examines color distributions and balance characteristics."
// }

// func generateAdvancedExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if advanced, exists := dataMap["advanced_assessment"]; exists {
// 			if assessment, ok := advanced.(map[string]interface{}); ok {
// 				if prob, exists := assessment["advanced_ai_probability"]; exists {
// 					if probability, ok := prob.(float64); ok {
// 						indicators := 0
// 						if advancedIndicators, exists := assessment["advanced_indicators"]; exists {
// 							if indicatorsMap, ok := advancedIndicators.(map[string]interface{}); ok {
// 								for _, value := range indicatorsMap {
// 									if boolValue, ok := value.(bool); ok && boolValue {
// 										indicators++
// 									}
// 								}
// 							}
// 						}

// 						if probability >= 0.6 {
// 							return fmt.Sprintf("Strong AI signals detected (%d indicators, %.0f%% probability). Advanced algorithms using Fourier transforms and pattern analysis.", indicators, probability*100)
// 						} else if indicators > 0 {
// 							return fmt.Sprintf("Weak AI signals detected (%d indicators, %.0f%% probability). Advanced mathematical analysis performed.", indicators, probability*100)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return "Advanced AI detection algorithms applied using specialized mathematical techniques."
// }

// func generateObjectCoherenceExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if objectAnalysis, exists := dataMap["object_analysis"]; exists {
// 			if analysis, ok := objectAnalysis.(map[string]interface{}); ok {
// 				if score, exists := analysis["ai_coherence_score"]; exists {
// 					if scoreFloat, ok := score.(float64); ok {
// 						if scoreFloat >= 0.6 {
// 							return fmt.Sprintf("Object analysis shows anomalies (%.0f%% AI probability). Found illogical size relationships or positioning.", scoreFloat*100)
// 						} else {
// 							return "Object analysis shows logical object arrangement and realistic size relationships."
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return "Object consistency analysis completed. Verifies logical plausibility of object arrangements."
// }

// func generateLightingExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if lightingAnalysis, exists := dataMap["lighting_analysis"]; exists {
// 			if analysis, ok := lightingAnalysis.(map[string]interface{}); ok {
// 				if score, exists := analysis["ai_lighting_score"]; exists {
// 					if scoreFloat, ok := score.(float64); ok {
// 						if scoreFloat >= 0.6 {
// 							return fmt.Sprintf("Lighting analysis reveals physical inconsistencies (%.0f%% AI probability). AI-generated images often have unrealistic lighting due to imperfect physics understanding.", scoreFloat*100)
// 						} else {
// 							return "Lighting analysis shows consistent lighting conditions following physical laws."
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return "Lighting analysis completed. Verifies consistency of light sources and shadows."
// }

// func generateAIModelExplanation(data interface{}) string {
// 	if dataMap, ok := data.(map[string]interface{}); ok {
// 		if prediction, exists := dataMap["prediction"]; exists {
// 			if predStr, ok := prediction.(string); ok {
// 				if probability, exists := dataMap["probability"]; exists {
// 					if probFloat, ok := probability.(float64); ok {
// 						if predStr == "fake" {
// 							return fmt.Sprintf("Deep learning model predicts AI-generated with %.1f%% confidence. This ResNet-18 model was specifically trained to distinguish AI from real images.", probFloat*100)
// 						} else {
// 							return fmt.Sprintf("Deep learning model predicts authentic with %.1f%% confidence. Neural network analysis suggests real photography.", probFloat*100)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return "Deep learning analysis completed using a trained neural network model."
// }

// // func generateMetadataExplanation(data interface{}) string {
// //     if dataMap, ok := data.(map[string]interface{}); ok {
// //         if len(dataMap) > 5 {
// //             return "Comprehensive metadata found containing technical information about image creation and file properties."
// //         } else if len(dataMap) > 0 {
// //             return "Basic metadata present with limited technical information."
// //         }
// //     }
// //     return "Metadata analysis completed."
// // }

// func calculateOverallVerdict(results []pythonrunner.ScriptResult) map[string]interface{} {
// 	scores := make(map[string]float64)
// 	var reasoning []string

// 	// Pipeline-Instanz für Score-Extraktion
// 	tempPipeline := analyzer.NewAnalysisPipeline()

// 	// Kategorisiere die Analysen
// 	traditionalMethods := []string{
// 		"artifacts", "compression", "pixel-analysis", "color-balance",
// 		"lighting-analysis", "advanced-artifacts", "object-coherence",
// 	}

// 	metadataMethods := []string{
// 		"metadata", "metadata-quick", "exif", "c2pa",
// 	}

// 	aiMethods := []string{
// 		"ai-model",
// 	}

// 	// Separate Score-Sammlung
// 	traditionalScores := make(map[string]float64)
// 	metadataScores := make(map[string]float64)
// 	aiScores := make(map[string]float64)

// 	// DEBUG: Log alle rohen Scores
// 	fmt.Printf("\n=== DEBUG SCORES ===\n")

// 	weights := map[string]float64{
// 		"metadata":           3.5,
// 		"c2pa":               3.5,
// 		"artifacts":          2.2,
// 		"lighting-analysis":  2.6,
// 		"advanced-artifacts": 2.0,
// 		"pixel-analysis":     1.8,
// 		"color-balance":      1.5,
// 		"object-coherence":   1.2,
// 		"compression":        2.5,
// 		"exif":               2.8,
// 		"ai-model":           4.0,
// 		"metadata-quick":     1.0,
// 	}

// 	var definitiveScore float64 = -1

// 	// 1. Prüfe auf DEFINITIVE Beweise und sammle kategorisierte Scores
// 	for _, result := range results {
// 		if result.Err != "" {
// 			continue
// 		}

// 		score := tempPipeline.ExtractConfidenceScore(result.Data)
// 		if score >= 0 {
// 			scores[result.Name] = score

// 			// Kategorisiere Score
// 			switch {
// 			case contains(traditionalMethods, result.Name):
// 				traditionalScores[result.Name] = score
// 			case contains(metadataMethods, result.Name):
// 				metadataScores[result.Name] = score
// 			case contains(aiMethods, result.Name):
// 				aiScores[result.Name] = score
// 			}

// 			// DEBUG: Log jeden Score
// 			fmt.Printf("RAW %s: %.3f\n", result.Name, score)

// 			// DEFINITIVE Beweise
// 			if result.Name == "metadata" && score >= 0.95 {
// 				definitiveScore = 1.0
// 				reasoning = append(reasoning, "Definitive AI metadata found")
// 				fmt.Printf("DEFINITIVE: Metadata AI found (%.3f)\n", score)
// 				break
// 			}
// 			if result.Name == "c2pa" && score >= 0.95 {
// 				definitiveScore = 1.0
// 				reasoning = append(reasoning, "C2PA certificate confirms AI generation")
// 				fmt.Printf("DEFINITIVE: C2PA AI found (%.3f)\n", score)
// 				break
// 			}
// 		}
// 	}

// 	// Berechne Kategorie-Durchschnitte
// 	traditionalAvg := calculateCategoryAverage(traditionalScores)
// 	metadataAvg := calculateCategoryAverage(metadataScores)
// 	aiAvg := calculateCategoryAverage(aiScores)

// 	// 2. Falls definitive Beweise: Return sofort
// 	if definitiveScore >= 0 {
// 		return map[string]interface{}{
// 			"verdict":          "AI Generated (Confirmed)",
// 			"probability":      100.0,
// 			"confidence":       0.99,
// 			"summary":          "AI Generated (Confirmed) - 100.0% AI probability with 99% confidence",
// 			"reasoning":        reasoning,
// 			"scores":           scores,
// 			"analysis_quality": calculateAnalysisQuality(results),
// 			"detailed_breakdown": map[string]interface{}{
// 				"traditional_computer_vision": map[string]interface{}{
// 					"average_score": traditionalAvg,
// 					"verdict":       getCategoryVerdict(traditionalAvg),
// 					"methods":       traditionalScores,
// 					"explanation":   generateTraditionalExplanation(traditionalScores),
// 				},
// 				"ai_deep_learning": map[string]interface{}{
// 					"average_score": aiAvg,
// 					"verdict":       getCategoryVerdict(aiAvg),
// 					"methods":       aiScores,
// 					"explanation":   generateAIExplanation(aiScores),
// 				},
// 				"metadata_forensics": map[string]interface{}{
// 					"average_score": metadataAvg,
// 					"verdict":       getCategoryVerdict(metadataAvg),
// 					"methods":       metadataScores,
// 					"explanation":   generateMetadataExplanation(metadataScores),
// 				},
// 				"method_agreement": analyzeMethodAgreement(traditionalAvg, aiAvg, metadataAvg),
// 			},
// 		}
// 	}

// 	// 3. Score-Kalibrierung
// 	calibratedScores := applyBalancedCalibration(scores)

// 	// DEBUG: Log kalibrierte Scores
// 	fmt.Printf("\n=== CALIBRATED SCORES ===\n")
// 	for name, score := range calibratedScores {
// 		fmt.Printf("CAL %s: %.3f (was %.3f)\n", name, score, scores[name])
// 	}

// 	// 4. Pattern-Boost
// 	patternBoost := calculateAdvancedBoost(calibratedScores)
// 	fmt.Printf("\nPATTERN BOOST: %.3f\n", patternBoost)

// 	// 5. Gewichtete Berechnung
// 	var weightedSum float64
// 	var totalWeight float64

// 	for _, result := range results {
// 		if result.Err != "" {
// 			continue
// 		}

// 		weight := weights[result.Name]
// 		score := calibratedScores[result.Name]

// 		if score >= 0 {
// 			// Moderate adaptive Gewichtung
// 			adaptiveWeight := weight
// 			if score >= 0.8 {
// 				adaptiveWeight *= 1.2
// 			} else if score <= 0.2 {
// 				adaptiveWeight *= 1.3
// 			}

// 			contribution := score * adaptiveWeight
// 			weightedSum += contribution
// 			totalWeight += adaptiveWeight

// 			// DEBUG: Log Beiträge
// 			fmt.Printf("CONTRIB %s: score=%.3f * weight=%.3f = %.3f\n",
// 				result.Name, score, adaptiveWeight, contribution)

// 			// Reasoning
// 			if score >= 0.7 {
// 				reasoning = append(reasoning, fmt.Sprintf("%s: Strong AI indicators (%.0f%% probability)", result.Name, score*100))
// 			} else if score <= 0.3 {
// 				reasoning = append(reasoning, fmt.Sprintf("%s: Authenticity indicators (%.0f%% authentic)", result.Name, (1-score)*100))
// 			} else {
// 				reasoning = append(reasoning, fmt.Sprintf("%s: Moderate signals (%.0f%% probability)", result.Name, score*100))
// 			}
// 		}
// 	}

// 	if totalWeight == 0 {
// 		return map[string]interface{}{
// 			"verdict":     "Analysis Failed",
// 			"probability": 0.0,
// 			"confidence":  0.0,
// 			"summary":     "No usable analysis results obtained",
// 			"reasoning":   []string{"Technical error during analysis"},
// 			"scores":      scores,
// 		}
// 	}

// 	// 6. Finale Score-Berechnung
// 	baseScore := weightedSum / totalWeight
// 	fmt.Printf("\nBASE SCORE: %.3f (weightedSum=%.3f / totalWeight=%.3f)\n",
// 		baseScore, weightedSum, totalWeight)

// 	baseScore *= patternBoost
// 	fmt.Printf("AFTER BOOST: %.3f\n", baseScore)

// 	// Qualitäts-Anpassung
// 	analysisQuality := float64(len(scores)) / 10.0
// 	qualityBonus := 1.0
// 	if analysisQuality >= 0.8 {
// 		qualityBonus = 1.05
// 	} else if analysisQuality < 0.5 {
// 		qualityBonus = 0.95
// 	}

// 	finalScore := baseScore * qualityBonus
// 	fmt.Printf("FINAL SCORE: %.3f (quality=%.3f)\n", finalScore, qualityBonus)

// 	// Clamp auf 0-1 Bereich
// 	if finalScore > 1.0 {
// 		finalScore = 1.0
// 	} else if finalScore < 0.0 {
// 		finalScore = 0.0
// 	}

// 	verdict, confidence := determineBalancedVerdict(finalScore, calibratedScores)

// 	fmt.Printf("VERDICT: %s (%.1f%%)\n", verdict, finalScore*100)
// 	fmt.Printf("==================\n\n")

// 	return map[string]interface{}{
// 		"verdict":          verdict,
// 		"probability":      finalScore * 100,
// 		"confidence":       confidence,
// 		"summary":          fmt.Sprintf("%s - %.0f AI points with %.0f%% confidence", verdict, finalScore*100, confidence*100),
// 		"reasoning":        reasoning,
// 		"scores":           scores,
// 		"analysis_quality": calculateAnalysisQuality(results),
// 		"quality_factor":   analysisQuality,

// 		"detailed_breakdown": map[string]interface{}{
// 			"traditional_computer_vision": map[string]interface{}{
// 				"average_score": traditionalAvg,
// 				"verdict":       getCategoryVerdict(traditionalAvg),
// 				"methods":       traditionalScores,
// 				"explanation":   generateTraditionalExplanation(traditionalScores),
// 			},
// 			"ai_deep_learning": map[string]interface{}{
// 				"average_score": aiAvg,
// 				"verdict":       getCategoryVerdict(aiAvg),
// 				"methods":       aiScores,
// 				"explanation":   generateAIExplanation(aiScores),
// 			},
// 			"metadata_forensics": map[string]interface{}{
// 				"average_score": metadataAvg,
// 				"verdict":       getCategoryVerdict(metadataAvg),
// 				"methods":       metadataScores,
// 				"explanation":   generateMetadataExplanation(metadataScores),
// 			},
// 			"method_agreement": analyzeMethodAgreement(traditionalAvg, aiAvg, metadataAvg),
// 		},
// 	}
// }

// // Hilfsfunktion für Array-Contains
// func contains(slice []string, item string) bool {
// 	for _, s := range slice {
// 		if s == item {
// 			return true
// 		}
// 	}
// 	return false
// }

// // Berechne Kategorie-Durchschnitt
// func calculateCategoryAverage(scores map[string]float64) float64 {
// 	if len(scores) == 0 {
// 		return -1 // Keine Daten verfügbar
// 	}

// 	var sum float64
// 	for _, score := range scores {
// 		sum += score
// 	}
// 	return sum / float64(len(scores))
// }

// // Kategorie-Verdict bestimmen
// func getCategoryVerdict(avgScore float64) string {
// 	if avgScore < 0 {
// 		return "No Data"
// 	} else if avgScore >= 0.7 {
// 		return "Strong AI Indicators"
// 	} else if avgScore >= 0.5 {
// 		return "Moderate AI Indicators"
// 	} else if avgScore >= 0.3 {
// 		return "Weak AI Indicators"
// 	} else {
// 		return "Authenticity Indicators"
// 	}
// }

// // Traditional Methods Erklärung
// func generateTraditionalExplanation(scores map[string]float64) string {
// 	highScores := []string{}
// 	lowScores := []string{}

// 	methodNames := map[string]string{
// 		"artifacts":          "Visual Artifacts",
// 		"compression":        "Compression Analysis",
// 		"pixel-analysis":     "Pixel-Level Analysis",
// 		"color-balance":      "Color Distribution",
// 		"lighting-analysis":  "Lighting Physics",
// 		"advanced-artifacts": "Mathematical Patterns",
// 		"object-coherence":   "Object Consistency",
// 	}

// 	for method, score := range scores {
// 		displayName := methodNames[method]
// 		if displayName == "" {
// 			displayName = method
// 		}

// 		if score >= 0.6 {
// 			highScores = append(highScores, displayName)
// 		} else if score <= 0.3 {
// 			lowScores = append(lowScores, displayName)
// 		}
// 	}

// 	if len(highScores) > 0 {
// 		return fmt.Sprintf("Computer Vision detected AI patterns in: %s", strings.Join(highScores, ", "))
// 	} else if len(lowScores) > 0 {
// 		return fmt.Sprintf("Computer Vision found authentic patterns in: %s", strings.Join(lowScores, ", "))
// 	} else {
// 		return "Computer Vision analysis shows mixed results"
// 	}
// }

// // AI Model Erklärung
// func generateAIExplanation(scores map[string]float64) string {
// 	if aiScore, exists := scores["ai-model"]; exists {
// 		if aiScore >= 0.8 {
// 			return fmt.Sprintf("Neural network strongly predicts AI-generated (%.1f%% confidence). Very high certainty from deep learning model.", aiScore*100)
// 		} else if aiScore >= 0.7 {
// 			return fmt.Sprintf("Neural network predicts AI-generated (%.1f%% confidence). High certainty from ResNet-18 model.", aiScore*100)
// 		} else if aiScore >= 0.5 {
// 			return fmt.Sprintf("Neural network leans towards AI-generated (%.1f%% confidence). Moderate certainty from trained model.", aiScore*100)
// 		} else if aiScore <= 0.2 {
// 			return fmt.Sprintf("Neural network strongly predicts authentic (%.1f%% confidence). Deep learning model very confident in real photography.", (1-aiScore)*100)
// 		} else if aiScore <= 0.3 {
// 			return fmt.Sprintf("Neural network predicts authentic (%.1f%% confidence). Model suggests real camera origin.", (1-aiScore)*100)
// 		} else {
// 			return fmt.Sprintf("Neural network uncertain (%.1f%% towards AI). Model shows mixed signals.", aiScore*100)
// 		}
// 	}
// 	return "No AI model analysis available"
// }

// // Metadata Erklärung
// func generateMetadataExplanation(scores map[string]float64) string {
// 	findings := []string{}

// 	for method, score := range scores {
// 		switch method {
// 		case "c2pa":
// 			if score >= 0.9 {
// 				findings = append(findings, "C2PA certificate confirms AI generation")
// 			} else if score <= 0.1 {
// 				findings = append(findings, "No C2PA AI markers found")
// 			} else {
// 				findings = append(findings, "C2PA analysis inconclusive")
// 			}
// 		case "exif":
// 			if score <= 0.1 {
// 				findings = append(findings, "Rich EXIF data suggests camera origin")
// 			} else if score >= 0.5 {
// 				findings = append(findings, "Suspicious EXIF patterns detected")
// 			} else {
// 				findings = append(findings, "Standard EXIF data present")
// 			}
// 		case "metadata":
// 			if score <= 0.2 {
// 				findings = append(findings, "Complete technical metadata present")
// 			} else if score >= 0.8 {
// 				findings = append(findings, "Metadata indicates AI generation")
// 			} else {
// 				findings = append(findings, "Standard metadata analysis")
// 			}
// 		case "metadata-quick":
// 			// Meist weniger relevant für Erklärung
// 			continue
// 		}
// 	}

// 	if len(findings) == 0 {
// 		return "Standard metadata analysis completed"
// 	}
// 	return strings.Join(findings, "; ")
// }

// // Method Agreement Analyse
// func analyzeMethodAgreement(traditional, ai, metadata float64) map[string]interface{} {
// 	agreement := map[string]interface{}{
// 		"overall_consensus":  "mixed",
// 		"conflicts":          []string{},
// 		"agreements":         []string{},
// 		"reliability_score":  0.5,
// 		"consensus_strength": "weak",
// 	}

// 	// Definiere Schwellenwerte
// 	highThreshold := 0.6
// 	lowThreshold := 0.3

// 	methods := map[string]float64{
// 		"Traditional Computer Vision": traditional,
// 		"AI Deep Learning":            ai,
// 		"Metadata Forensics":          metadata,
// 	}

// 	// Zähle High/Low Scores
// 	highCount := 0
// 	lowCount := 0
// 	highMethods := []string{}
// 	lowMethods := []string{}

// 	for name, score := range methods {
// 		if score >= 0 { // Only count methods that have data
// 			if score >= highThreshold {
// 				highCount++
// 				highMethods = append(highMethods, name)
// 			} else if score <= lowThreshold {
// 				lowCount++
// 				lowMethods = append(lowMethods, name)
// 			}
// 		}
// 	}

// 	// Analysiere Konsens
// 	totalMethods := 0
// 	for _, score := range methods {
// 		if score >= 0 {
// 			totalMethods++
// 		}
// 	}

// 	if highCount >= 2 && totalMethods >= 2 {
// 		agreement["overall_consensus"] = "ai_likely"
// 		agreement["reliability_score"] = 0.8 + (float64(highCount) * 0.05)
// 		agreement["consensus_strength"] = "strong"
// 		for _, method := range highMethods {
// 			agreement["agreements"] = append(agreement["agreements"].([]string),
// 				fmt.Sprintf(" %s detects AI indicators", method))
// 		}
// 	} else if lowCount >= 2 && totalMethods >= 2 {
// 		agreement["overall_consensus"] = "authentic_likely"
// 		agreement["reliability_score"] = 0.8 + (float64(lowCount) * 0.05)
// 		agreement["consensus_strength"] = "strong"
// 		for _, method := range lowMethods {
// 			agreement["agreements"] = append(agreement["agreements"].([]string),
// 				fmt.Sprintf(" %s finds authenticity signs", method))
// 		}
// 	} else {
// 		agreement["overall_consensus"] = "conflicting"
// 		agreement["reliability_score"] = 0.3 + (float64(totalMethods) * 0.1)
// 		agreement["consensus_strength"] = "weak"

// 		// Identifiziere spezifische Konflikte
// 		if traditional >= highThreshold && ai >= 0 && ai <= lowThreshold {
// 			agreement["conflicts"] = append(agreement["conflicts"].([]string),
// 				"Computer Vision detects AI, but Deep Learning suggests authentic")
// 		}
// 		if ai >= highThreshold && traditional >= 0 && traditional <= lowThreshold {
// 			agreement["conflicts"] = append(agreement["conflicts"].([]string),
// 				"Deep Learning detects AI, but Computer Vision suggests authentic")
// 		}
// 		if metadata >= 0 && metadata <= lowThreshold && (traditional >= highThreshold || ai >= highThreshold) {
// 			agreement["conflicts"] = append(agreement["conflicts"].([]string),
// 				"Clean metadata conflicts with analysis results")
// 		}
// 		if metadata >= highThreshold && (traditional <= lowThreshold || ai <= lowThreshold) {
// 			agreement["conflicts"] = append(agreement["conflicts"].([]string),
// 				"Suspicious metadata conflicts with analysis results")
// 		}

// 		// Füge positive Übereinstimmungen hinzu auch bei Konflikten
// 		if len(highMethods) > 0 {
// 			for _, method := range highMethods {
// 				agreement["agreements"] = append(agreement["agreements"].([]string),
// 					fmt.Sprintf("%s shows AI indicators", method))
// 			}
// 		}
// 		if len(lowMethods) > 0 {
// 			for _, method := range lowMethods {
// 				agreement["agreements"] = append(agreement["agreements"].([]string),
// 					fmt.Sprintf("%s shows authenticity signs", method))
// 			}
// 		}
// 	}

// 	return agreement
// }

// func applyBalancedCalibration(scores map[string]float64) map[string]float64 {
// 	calibratedScores := make(map[string]float64)

// 	calibrationFactors := map[string]float64{
// 		"lighting-analysis":  1.30, // Gut für AI-Erkennung
// 		"artifacts":          1.05, // Weniger aggressiv bei PNG
// 		"advanced-artifacts": 1.15, // Mathematische Analyse
// 		"pixel-analysis":     1.05, // Weniger aggressiv bei echten Fotos
// 		"color-balance":      1.00, // Neutral - sehr zuverlässig
// 		"object-coherence":   0.90, // Leicht reduziert
// 		"compression":        0.40, // Starker Authentizitäts-Boost
// 		"metadata":           1.0,  // Sehr zuverlässig
// 		"c2pa":               1.0,  // Sehr zuverlässig
// 		"exif":               0.70, // Authentizitäts-Boost
// 		"ai-model":           1.0,  // Neutral - wird dynamisch angepasst
// 		"metadata-quick":     1.0,  // Standard
// 	}

// 	for name, score := range scores {
// 		if factor, exists := calibrationFactors[name]; exists {
// 			calibratedScore := score * factor
// 			calibratedScores[name] = math.Min(1.0, calibratedScore)
// 		} else {
// 			calibratedScores[name] = score
// 		}
// 	}

// 	return calibratedScores
// }

// func calculateAdvancedBoost(scores map[string]float64) float64 {
// 	boost := 1.0

// 	// Pattern 1: Konsistenz zwischen traditionellen Methoden
// 	consistencyCount := 0
// 	highScoreCount := 0

// 	traditionalMethods := []string{"artifacts", "lighting-analysis", "advanced-artifacts", "pixel-analysis"}
// 	for _, method := range traditionalMethods {
// 		if score, exists := scores[method]; exists {
// 			if score >= 0.6 {
// 				highScoreCount++
// 			}
// 			if score >= 0.4 {
// 				consistencyCount++
// 			}
// 		}
// 	}

// 	// Starke Konsistenz = höherer Boost
// 	if consistencyCount >= 3 {
// 		boost *= 1.4
// 	} else if consistencyCount >= 2 {
// 		boost *= 1.2
// 	}

// 	// Pattern 2: AI-Model + Traditional Methods Alignment
// 	if aiModelScore, exists := scores["ai-model"]; exists {
// 		if artifactsScore, exists := scores["artifacts"]; exists {
// 			// Beide Methoden erkennen AI-Indizien
// 			if aiModelScore >= 0.5 && artifactsScore >= 0.6 {
// 				boost *= 1.5 // Starker Boost bei Übereinstimmung
// 			}
// 			// AI-Model erkennt AI, aber Artifacts nicht
// 			if aiModelScore >= 0.7 && artifactsScore <= 0.3 {
// 				boost *= 1.3 // AI-Model überstimmt traditionelle Methoden
// 			}
// 		}

// 		if lightingScore, exists := scores["lighting-analysis"]; exists {
// 			// AI-Model + Lighting Analysis Alignment
// 			if aiModelScore >= 0.5 && lightingScore >= 0.4 {
// 				boost *= 1.3
// 			}
// 		}

// 		// Pattern: AI-Model sagt "authentisch" + niedrige traditionelle Scores
// 		if aiModelScore <= 0.3 { // AI-Modell ist sehr sicher dass es authentisch ist
// 			authenticityCount := 0
// 			if compressionScore, exists := scores["compression"]; exists && compressionScore <= 0.3 {
// 				authenticityCount++
// 			}
// 			if colorScore, exists := scores["color-balance"]; exists && colorScore <= 0.2 {
// 				authenticityCount++
// 			}
// 			if artifactsScore, exists := scores["artifacts"]; exists && artifactsScore <= 0.4 {
// 				authenticityCount++
// 			}

// 			if authenticityCount >= 2 {
// 				boost *= 0.6 // Starker Authentizitäts-Boost
// 			}
// 		}

// 		// Pattern: AI-Model sehr sicher (>80%) - verstärke dessen Einfluss
// 		if aiModelScore >= 0.8 || aiModelScore <= 0.2 {
// 			boost *= 1.4 // Hohe Confidence des AI-Models verstärken
// 		}
// 	}

// 	// Pattern 3: Lighting + Physics Konsistenz
// 	if lightingScore, exists := scores["lighting-analysis"]; exists {
// 		if objectScore, exists := scores["object-coherence"]; exists {
// 			if lightingScore >= 0.6 && objectScore >= 0.5 {
// 				boost *= 1.3 // Physik-Inkonsistenzen verstärken sich
// 			}
// 		}
// 	}

// 	// Pattern 4: Compression + Metadata Diskrepanz
// 	if compressionScore, exists := scores["compression"]; exists {
// 		if metadataScore, exists := scores["metadata"]; exists {
// 			// Hohe Compression-Anomalien aber saubere Metadata = verdächtig
// 			if compressionScore >= 0.6 && metadataScore <= 0.2 {
// 				boost *= 1.2
// 			}
// 		}
// 	}

// 	// Pattern 5: Advanced Artifacts + Color Balance
// 	if advancedScore, exists := scores["advanced-artifacts"]; exists {
// 		if colorScore, exists := scores["color-balance"]; exists {
// 			if advancedScore >= 0.5 && colorScore >= 0.4 {
// 				boost *= 1.25 // Mathematische + Farbanomalien
// 			}
// 		}
// 	}

// 	// Pattern 6: Viele niedrige Scores = wahrscheinlich authentisch
// 	authenticityIndicators := 0
// 	totalIndicators := 0

// 	for method, score := range scores {
// 		if method != "metadata" && method != "exif" && method != "metadata-quick" {
// 			totalIndicators++
// 			if score <= 0.3 {
// 				authenticityIndicators++
// 			}
// 		}
// 	}

// 	if totalIndicators >= 5 && authenticityIndicators >= 4 {
// 		boost *= 0.85 // Starke Authentizitäts-Indizien
// 	}

// 	// Pattern 7: Pixel-Level + Advanced Artifacts Kombination
// 	if pixelScore, exists := scores["pixel-analysis"]; exists {
// 		if advancedScore, exists := scores["advanced-artifacts"]; exists {
// 			if pixelScore >= 0.6 && advancedScore >= 0.5 {
// 				boost *= 1.3 // Mathematische Anomalien auf verschiedenen Ebenen
// 			}
// 		}
// 	}

// 	// Pattern 8: Metadata-Mangel + hohe AI-Scores
// 	metadataPresent := false
// 	if metadataScore, exists := scores["metadata"]; exists && metadataScore <= 0.1 {
// 		metadataPresent = true
// 	}
// 	if exifScore, exists := scores["exif"]; exists && exifScore <= 0.1 {
// 		metadataPresent = true
// 	}

// 	if !metadataPresent {
// 		// Fehlende Metadata + hohe AI-Scores = verdächtig
// 		if aiModelScore, exists := scores["ai-model"]; exists && aiModelScore >= 0.5 {
// 			boost *= 1.2
// 		}
// 		if artifactsScore, exists := scores["artifacts"]; exists && artifactsScore >= 0.5 {
// 			boost *= 1.15
// 		}
// 	}

// 	// Begrenze den Boost-Faktor
// 	if boost > 2.0 {
// 		boost = 2.0
// 	} else if boost < 0.4 {
// 		boost = 0.4
// 	}

// 	return boost
// }

// // Helper-Funktion für EXIF-Erkennung
// func hasExifData(scores map[string]float64) bool {
// 	exifScore, exists := scores["exif"]
// 	if !exists {
// 		return true // Default: EXIF verfügbar
// 	}

// 	// Wenn EXIF-Score niedrig, sind EXIF-Daten vorhanden
// 	return exifScore <= 0.3
// }

// func determineBalancedVerdict(score float64, scores map[string]float64) (string, float64) {
// 	// Score-Verteilung analysieren
// 	highConfidenceScores := 0
// 	lowConfidenceScores := 0

// 	for _, individualScore := range scores {
// 		if individualScore >= 0.7 {
// 			highConfidenceScores++
// 		} else if individualScore <= 0.3 {
// 			lowConfidenceScores++
// 		}
// 	}

// 	// Base confidence
// 	totalScores := len(scores)
// 	baseConfidence := 0.6

// 	if totalScores > 0 {
// 		definitiveRatio := float64(highConfidenceScores+lowConfidenceScores) / float64(totalScores)
// 		baseConfidence = 0.5 + (definitiveRatio * 0.4)
// 	}

// 	// Threshold adjustments basierend auf Score-Verteilung
// 	thresholdAdjustment := 0.0
// 	if highConfidenceScores >= 3 {
// 		thresholdAdjustment = -0.10 // Niedrigere Schwelle bei vielen hohen Scores
// 	} else if lowConfidenceScores >= 3 {
// 		thresholdAdjustment = 0.12 // Höhere Schwelle bei vielen niedrigen Scores
// 	}

// 	// Verdict-Bestimmung mit angepassten Schwellenwerten
// 	if score >= (0.45 + thresholdAdjustment) {
// 		confidence := math.Min(0.95, baseConfidence+0.25)
// 		return "Very Likely AI Generated", confidence
// 	} else if score >= (0.30 + thresholdAdjustment) {
// 		confidence := math.Min(0.9, baseConfidence+0.15)
// 		return "Likely AI Generated", confidence
// 	} else if score >= (0.22 + thresholdAdjustment) {
// 		confidence := math.Min(0.8, baseConfidence)
// 		return "Possibly AI Generated", confidence
// 	} else if score >= (0.16 + thresholdAdjustment) {
// 		confidence := math.Min(0.85, baseConfidence+0.1)
// 		return "Likely Authentic", confidence
// 	} else {
// 		confidence := math.Min(0.9, baseConfidence+0.2)
// 		return "Very Likely Authentic", confidence
// 	}
// }

// func calculateAnalysisQuality(results []pythonrunner.ScriptResult) string {
// 	successfulAnalyses := 0
// 	totalAnalyses := len(results)

// 	for _, result := range results {
// 		if result.Err == "" {
// 			successfulAnalyses++
// 		}
// 	}

// 	if totalAnalyses == 0 {
// 		return "no_data"
// 	}

// 	successRate := float64(successfulAnalyses) / float64(totalAnalyses)

// 	if successRate >= 0.9 {
// 		return "excellent"
// 	} else if successRate >= 0.7 {
// 		return "good"
// 	} else if successRate >= 0.5 {
// 		return "fair"
// 	} else {
// 		return "poor"
// 	}
// }

// func metricsHandler(c *gin.Context) {
// 	fmt.Printf("DEBUG: metrics variable: %+v\n", metrics)

// 	m := metrics

// 	if m == nil {
// 		fmt.Printf("DEBUG: metrics is nil!\n")
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "metrics not initialized"})
// 		return
// 	}

// 	response := m.GetMetricsSummary()

// 	fmt.Printf("DEBUG: Full response: %+v\n", response)

// 	response["timestamp"] = time.Now().Unix()

// 	c.Header("Content-Type", "application/json")
// 	c.JSON(http.StatusOK, response)
// }

// func getFloatValue(m map[string]interface{}, key string, defaultValue float64) float64 {
// 	if val, ok := m[key]; ok {
// 		if floatVal, ok := val.(float64); ok {
// 			return floatVal
// 		}
// 	}
// 	return defaultValue
// }

// func getInt64Value(m map[string]interface{}, key string, defaultValue int64) int64 {
// 	if val, ok := m[key]; ok {
// 		if intVal, ok := val.(int64); ok {
// 			return intVal
// 		}
// 		if intVal, ok := val.(int); ok {
// 			return int64(intVal)
// 		}
// 	}
// 	return defaultValue
// }

// // Health Check Handler
// func healthHandler(c *gin.Context) {
// 	summary := metrics.GetMetricsSummary()

// 	healthy := true
// 	issues := []string{}

// 	if business, ok := summary["business"].(map[string]interface{}); ok {
// 		if errorRate, ok := business["error_rate"].(float64); ok && errorRate > 0.1 {
// 			healthy = false
// 			issues = append(issues, "High error rate detected")
// 		}
// 	}

// 	status := http.StatusOK
// 	if !healthy {
// 		status = http.StatusServiceUnavailable
// 	}

// 	c.JSON(status, gin.H{
// 		"healthy":   healthy,
// 		"issues":    issues,
// 		"metrics":   summary,
// 		"timestamp": time.Now().Unix(),
// 	})
// }

// func RegisterHandlers(r *gin.Engine, logger *logging.Logger) *monitoring.Metrics {
// 	SetGlobalLogger(logger)
// 	r.POST("/upload", uploadHandler)
// 	r.GET("/metrics", metricsHandler)
// 	r.GET("/health", healthHandler)
// 	return metrics
// }
