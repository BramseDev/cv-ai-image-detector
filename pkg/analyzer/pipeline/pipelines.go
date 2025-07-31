package pipeline

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BramseDev/imageAnalyzer/cache"
	exifanalyzer "github.com/BramseDev/imageAnalyzer/pkg/analyzer/exif_analyzer"
	"github.com/BramseDev/imageAnalyzer/pkg/pythonrunner"
	"github.com/BramseDev/imageAnalyzer/pkg/rustrunner"
)

// AnalysisStage definiert eine einzelne Analysestufe
type AnalysisStage struct {
	Name         string
	Priority     int
	FastTrack    bool
	Timeout      time.Duration
	Dependencies []string
	Analyzer     func(context.Context, string) (interface{}, error)
}

// PipelineResult enthält das Gesamtergebnis der Pipeline
type PipelineResult struct {
	Results     map[string]interface{}
	StagesRun   []string
	ProcessTime time.Duration
	EarlyExit   bool
	Confidence  float64
	CacheHit    bool
}

var (
	globalCache *cache.AnalysisCache
	cacheOnce   sync.Once
)

func getGlobalCache() *cache.AnalysisCache {
	cacheOnce.Do(func() {
		globalCache = cache.NewAnalysisCache()
	})
	return globalCache
}

// MetricsRecorder Interface für Cache-Tracking
type MetricsRecorder interface {
	RecordCacheHit()
	RecordCacheMiss()
}

// AnalysisPipeline Hauptstruktur
type AnalysisPipeline struct {
	stages           []AnalysisStage
	cache            *cache.AnalysisCache
	metrics          MetricsRecorder
	earlyExitEnabled bool
	mu               sync.RWMutex
}

// NewAnalysisPipeline erstellt eine neue Pipeline-Instanz ohne Cache
func NewAnalysisPipeline() *AnalysisPipeline {
	return &AnalysisPipeline{
		stages:           getDefaultStages(),
		cache:            getGlobalCache(),
		earlyExitEnabled: true,
	}
}

// NewAnalysisPipelineWithCache erstellt eine Pipeline mit Metrics-Integration
func NewAnalysisPipelineWithCache(metrics MetricsRecorder) *AnalysisPipeline {
	return &AnalysisPipeline{
		stages:           getDefaultStages(),
		cache:            getGlobalCache(),
		metrics:          metrics,
		earlyExitEnabled: true,
	}
}

// getDefaultStages definiert alle verfügbaren Analysestufen
func getDefaultStages() []AnalysisStage {
	return []AnalysisStage{
		// Priorität 1: Schnelle, definitive Checks
		{
			Name:         "metadata-quick",
			Priority:     1,
			FastTrack:    true,
			Timeout:      5 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunMetadata,
		},
		{
			Name:         "c2pa",
			Priority:     1,
			FastTrack:    true,
			Timeout:      8 * time.Second,
			Dependencies: []string{},
			Analyzer:     rustrunner.RunC2PA,
		},
		{
			Name:      "exif",
			Priority:  1,
			FastTrack: true,
			Timeout:   2 * time.Second,
			Analyzer: func(ctx context.Context, p string) (interface{}, error) {
				return exifanalyzer.AnalyzeEXIF(p)
			},
		},

		// Priorität 2: Wichtige technische Analysen
		{
			Name:         "metadata",
			Priority:     2,
			FastTrack:    false,
			Timeout:      8 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunMetadata,
		},

		// Priorität 3: Spezialisierte Bildanalysen
		{
			Name:         "artifacts",
			Priority:     3,
			FastTrack:    false,
			Timeout:      15 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunArtifacts,
		},
		{
			Name:         "compression",
			Priority:     3,
			FastTrack:    false,
			Timeout:      10 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunCompression,
		},
		{
			Name:         "pixel-analysis",
			Priority:     3,
			FastTrack:    false,
			Timeout:      18 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunPixelAnalysis,
		},
		{
			Name:         "color-balance",
			Priority:     3,
			FastTrack:    false,
			Timeout:      12 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunColorBalance,
		},
		{
			Name:         "advanced-artifacts",
			Priority:     3,
			FastTrack:    false,
			Timeout:      20 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunAdvancedArtifacts,
		},

		// Priorität 4: Neue visuelle Inhaltanalysen
		{
			Name:         "object-coherence",
			Priority:     4,
			FastTrack:    false,
			Timeout:      25 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunObjectCoherence,
		},
		{
			Name:         "lighting-analysis",
			Priority:     4,
			FastTrack:    false,
			Timeout:      20 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunLightingAnalysis,
		},
		{
			Name:         "ai-model",
			Priority:     2,
			FastTrack:    false,
			Timeout:      30 * time.Second,
			Dependencies: []string{},
			Analyzer:     pythonrunner.RunAIModelPrediction,
		},
	}
}

// SetEarlyExitEnabled aktiviert/deaktiviert Early Exit
func (ap *AnalysisPipeline) SetEarlyExitEnabled(enabled bool) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.earlyExitEnabled = enabled
}

// IsEarlyExitEnabled prüft ob Early Exit aktiviert ist
func (ap *AnalysisPipeline) IsEarlyExitEnabled() bool {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.earlyExitEnabled
}

// ExtractConfidenceScore extrahiert Score (exportiert für upload.go)
func (ap *AnalysisPipeline) ExtractConfidenceScore(data interface{}) float64 {
	return ap.extractConfidenceScore(data)
}

func (ap *AnalysisPipeline) RunAnalysis(ctx context.Context, imagePath string) (*PipelineResult, error) {
	startTime := time.Now()
	logger := slog.With("image_path", imagePath)

	// Cache-Key generieren
	cacheKey, err := ap.generateCacheKey(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cache key: %w", err)
	}

	logger = logger.With("cache_key", cacheKey[:16])

	// DEBUG: Cache-Status prüfen
	logger.Info("DEBUG: Checking cache",
		"cache_enabled", ap.cache != nil,
		"cache_key", cacheKey[:16])

	// Cache prüfen
	if cachedResult, found := ap.cache.Get(cacheKey); found {
		if ap.metrics != nil {
			ap.metrics.RecordCacheHit()
		}

		if result, ok := cachedResult.(*PipelineResult); ok {
			cachedCopy := *result
			cachedCopy.CacheHit = true
			cachedCopy.ProcessTime = time.Since(startTime)

			logger.Info("Cache HIT - returning cached result",
				"original_duration", result.ProcessTime,
				"cache_lookup_time", time.Since(startTime))

			return &cachedCopy, nil
		} else {
			logger.Warn("Cache entry found but wrong type", "type", fmt.Sprintf("%T", cachedResult))
		}
	} else {
		logger.Info("Cache entry not found or expired")
	}

	// Cache Miss - Record it
	if ap.metrics != nil {
		ap.metrics.RecordCacheMiss()
	}

	logger.Info("Cache MISS - running full analysis")

	// Rest bleibt gleich...
	result, err := ap.runFullAnalysis(ctx, imagePath, logger)
	if err != nil {
		return nil, err
	}

	result.CacheHit = false
	result.ProcessTime = time.Since(startTime)

	// DEBUG: Cache speichern
	logger.Info("DEBUG: Storing in cache",
		"cache_key", cacheKey[:16],
		"result_type", fmt.Sprintf("%T", result))

	// Speichere Ergebnis im Cache
	ap.cache.Set(cacheKey, result, 30*time.Minute)

	logger.Info("Analysis completed and cached",
		"duration", result.ProcessTime,
		"stages_completed", len(result.StagesRun))

	return result, nil
}

func (ap *AnalysisPipeline) generateCacheKey(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Cache-Key nur basierend auf Inhalt, nicht auf Pfad
	hashStr := fmt.Sprintf("%x", hash.Sum(nil))
	return fmt.Sprintf("analysis_%s", hashStr), nil
}

// runFullAnalysis führt die tatsächliche Analyse ohne Cache aus
func (ap *AnalysisPipeline) runFullAnalysis(ctx context.Context, imagePath string, logger *slog.Logger) (*PipelineResult, error) {
	result := &PipelineResult{
		Results:   make(map[string]interface{}),
		StagesRun: []string{},
		EarlyExit: false,
	}

	// Sortiere Stages nach Priorität
	stages := ap.getSortedStages()

	// Phase 1: FastTrack Stages (für Early Exit)
	if ap.IsEarlyExitEnabled() {
		fastTrackResults := ap.runFastTrackStages(ctx, imagePath, stages, logger)

		// Merge FastTrack Ergebnisse
		for name, data := range fastTrackResults {
			result.Results[name] = data
			result.StagesRun = append(result.StagesRun, name)
		}

		// Prüfe Early Exit Bedingungen
		if ap.shouldEarlyExit(result.Results) {
			result.EarlyExit = true
			result.Confidence = ap.calculateEarlyConfidence(result.Results)

			logger.Info("Early exit triggered",
				"stages_run", len(result.StagesRun),
				"confidence", result.Confidence)
			return result, nil
		}
	}

	// Phase 2: Vollständige Analyse
	logger.Info("Running full analysis pipeline")

	for _, stage := range stages {
		// Skip bereits ausgeführte FastTrack Stages
		if stage.FastTrack && result.Results[stage.Name] != nil {
			continue
		}

		// Context mit Stage-spezifischem Timeout
		stageCtx, cancel := context.WithTimeout(ctx, stage.Timeout)

		logger.Info("Running stage", "stage", stage.Name, "timeout", stage.Timeout)
		stageStart := time.Now()

		stageResult, err := stage.Analyzer(stageCtx, imagePath)
		stageDuration := time.Since(stageStart)

		cancel() // Cleanup

		if err != nil {
			logger.Warn("Stage failed",
				"stage", stage.Name,
				"error", err,
				"duration", stageDuration)

			// Nicht-kritische Fehler: Weiter
			continue
		}

		result.Results[stage.Name] = stageResult
		result.StagesRun = append(result.StagesRun, stage.Name)

		logger.Info("Stage completed",
			"stage", stage.Name,
			"duration", stageDuration)

		// Prüfe Context Cancellation
		select {
		case <-ctx.Done():
			logger.Warn("Pipeline cancelled", "completed_stages", len(result.StagesRun))
			return nil, ctx.Err()
		default:
			// Continue
		}
	}

	// Finale Berechnung
	result.Confidence = ap.calculateFinalConfidence(result.Results)

	logger.Info("Pipeline completed",
		"total_stages", len(result.StagesRun),
		"confidence", result.Confidence)

	return result, nil
}

// Helper methods...
func (ap *AnalysisPipeline) getSortedStages() []AnalysisStage {
	stages := make([]AnalysisStage, len(ap.stages))
	copy(stages, ap.stages)

	// Bubble Sort nach Priorität
	for i := 0; i < len(stages)-1; i++ {
		for j := 0; j < len(stages)-i-1; j++ {
			if stages[j].Priority > stages[j+1].Priority {
				stages[j], stages[j+1] = stages[j+1], stages[j]
			}
		}
	}
	return stages
}

func (ap *AnalysisPipeline) runFastTrackStages(ctx context.Context, imagePath string, stages []AnalysisStage, logger *slog.Logger) map[string]interface{} {
	results := make(map[string]interface{})

	for _, stage := range stages {
		if !stage.FastTrack {
			continue
		}

		stageCtx, cancel := context.WithTimeout(ctx, stage.Timeout)
		stageResult, err := stage.Analyzer(stageCtx, imagePath)
		cancel()

		if err != nil {
			logger.Warn("FastTrack stage failed", "stage", stage.Name, "error", err)
			continue
		}

		results[stage.Name] = stageResult
		logger.Info("FastTrack stage completed", "stage", stage.Name)
	}
	return results
}

func (ap *AnalysisPipeline) shouldEarlyExit(results map[string]interface{}) bool {
	if !ap.IsEarlyExitEnabled() {
		return false
	}

	if metaResult, exists := results["metadata-quick"]; exists {
		if ap.hasDefinitiveMetadataEvidence(metaResult) {
			return true
		}
	}

	if c2paResult, exists := results["c2pa"]; exists {
		if ap.hasDefinitiveC2PAEvidence(c2paResult) {
			return true
		}
	}

	return false
}

func (ap *AnalysisPipeline) hasDefinitiveMetadataEvidence(data interface{}) bool {
	if dataMap, ok := data.(map[string]interface{}); ok {
		return ap.detectAIFromMetadata(dataMap) >= 0.95
	}
	return false
}

func (ap *AnalysisPipeline) hasDefinitiveC2PAEvidence(data interface{}) bool {
	if dataMap, ok := data.(map[string]interface{}); ok {
		if score, exists := dataMap["score"]; exists {
			if scoreFloat, ok := score.(float64); ok {
				return scoreFloat >= 95
			}
		}
	}
	return false
}

func (ap *AnalysisPipeline) calculateEarlyConfidence(results map[string]interface{}) float64 {
	return 0.98 // Hohe Konfidenz bei Early Exit
}

func (ap *AnalysisPipeline) calculateFinalConfidence(results map[string]interface{}) float64 {
	totalStages := len(ap.stages)
	completedStages := len(results)

	completionRatio := float64(completedStages) / float64(totalStages)
	return 0.5 + (completionRatio * 0.4) // 0.5 - 0.9
}

func (ap *AnalysisPipeline) extractConfidenceScore(data interface{}) float64 {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return -1
	}

	// Color-Balance Score
	if aiColorScore, exists := dataMap["ai_color_score"]; exists {
		if score, ok := aiColorScore.(float64); ok {
			return score
		}
	}

	// Advanced Artifacts Score
	if advanced, exists := dataMap["advanced_assessment"]; exists {
		if assessment, ok := advanced.(map[string]interface{}); ok {
			if prob, exists := assessment["advanced_ai_probability"]; exists {
				if probability, ok := prob.(float64); ok {
					return probability
				}
			}
		}
	}

	// Compression Analysis Score
	if compressionAnalysis, exists := dataMap["compression_ai_analysis"]; exists {
		if analysis, ok := compressionAnalysis.(map[string]interface{}); ok {
			if prob, exists := analysis["ai_probability"]; exists {
				if probFloat, ok := prob.(float64); ok {
					return probFloat
				}
			}
		}
	}

	// Für nested compression data
	for _, value := range dataMap {
		if fileData, ok := value.(map[string]interface{}); ok {
			if analysis, exists := fileData["compression_ai_analysis"]; exists {
				if analysisMap, ok := analysis.(map[string]interface{}); ok {
					if prob, exists := analysisMap["ai_probability"]; exists {
						if probFloat, ok := prob.(float64); ok {
							return probFloat
						}
					}
				}
			}
		}
	}

	// Artifacts Score
	if overall, exists := dataMap["overall_assessment"]; exists {
		if assessment, ok := overall.(map[string]interface{}); ok {
			if score, exists := assessment["ai_probability_score"]; exists {
				if scoreFloat, ok := score.(float64); ok {
					return scoreFloat
				}
			}
		}
	}

	// C2PA Score
	if c2paScore, exists := dataMap["score"]; exists {
		if scoreFloat, ok := c2paScore.(float64); ok {
			return scoreFloat / 100.0 // Normalisiere auf 0-1
		}
	}

	// Pixel Analysis Score
	if pixelOverall, exists := dataMap["overall_assessment"]; exists {
		if assessment, ok := pixelOverall.(map[string]interface{}); ok {
			if score, exists := assessment["ai_probability_score"]; exists {
				if scoreFloat, ok := score.(float64); ok {
					return scoreFloat
				}
			}
		}
	}

	// Lighting Analysis Score
	if lightingAnalysis, exists := dataMap["lighting_analysis"]; exists {
		if analysis, ok := lightingAnalysis.(map[string]interface{}); ok {
			if score, exists := analysis["ai_lighting_score"]; exists {
				if scoreFloat, ok := score.(float64); ok {
					return scoreFloat
				}
			}
		}
	}

	// Object Coherence Score
	if objectAnalysis, exists := dataMap["object_analysis"]; exists {
		if analysis, ok := objectAnalysis.(map[string]interface{}); ok {
			if score, exists := analysis["ai_coherence_score"]; exists {
				if scoreFloat, ok := score.(float64); ok {
					return scoreFloat
				}
			}
		}
	}
	if prediction, exists := dataMap["prediction"]; exists {
		if predStr, ok := prediction.(string); ok {
			if probability, exists := dataMap["probability"]; exists {
				if probFloat, ok := probability.(float64); ok {
					if predStr == "fake" {
						return probFloat
					} else {
						return 1.0 - probFloat
					}
				}
			}
		}
	}

	return -1
}

func (ap *AnalysisPipeline) detectAIFromMetadata(dataMap map[string]interface{}) float64 {
	aiKeywords := []string{
		"ChatGPT", "DALL-E", "Midjourney", "Stable Diffusion",
		"trainedAlgorithmicMedia", "AI generated",
	}

	for _, value := range dataMap {
		if valueStr, ok := value.(string); ok {
			for _, keyword := range aiKeywords {
				if strings.Contains(strings.ToLower(valueStr), strings.ToLower(keyword)) {
					return 1.0
				}
			}
		}
	}
	return -1
}

// Cache Management Methods
func (ap *AnalysisPipeline) ClearCache() {
	if ap.cache != nil {
		// Implementiere Cache-Clear Methode falls nötig
		ap.cache = cache.NewAnalysisCache()
	}
}

// GetCacheStats mit echten Informationen
func (ap *AnalysisPipeline) GetCacheStats() map[string]interface{} {
	if ap.cache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	// Versuche Cache-Statistiken zu ermitteln
	stats := map[string]interface{}{
		"enabled": true,
	}

	// Falls dein Cache-Package Stats unterstützt, füge sie hinzu
	// stats["entries"] = ap.cache.Len()
	// stats["hit_rate"] = ap.cache.HitRate()

	return stats
}
