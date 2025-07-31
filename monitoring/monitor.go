package monitoring

import (
	"fmt"
	"sync"
	"time"

	analyzer "github.com/BramseDev/imageAnalyzer/pkg/analyzer/pipeline"
)

type Metrics struct {
	mu sync.RWMutex

	// Performance Metrics
	AnalysisCount    map[string]int64
	AnalysisDuration map[string][]time.Duration
	ErrorCount       map[string]int64

	// Business Metrics - FIXED
	TotalAnalyses   int64
	AIDetectedCount int64
	EarlyExitCount  int64
	AIDetectionRate float64
	EarlyExitRate   float64
	CacheHitRate    float64

	// System Metrics
	MemoryUsage       []float64
	CPUUsage          []float64
	ActiveConnections int64

	// Cache Metrics
	CacheHits   int64
	CacheMisses int64

	// Additional tracking
	LastUpdate time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		AnalysisCount:    make(map[string]int64),
		AnalysisDuration: make(map[string][]time.Duration),
		ErrorCount:       make(map[string]int64),
		MemoryUsage:      make([]float64, 0),
		CPUUsage:         make([]float64, 0),
		LastUpdate:       time.Now(),
	}
}

// Record analysis results properly
func (m *Metrics) RecordAnalysisResult(isAI bool, isEarlyExit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// m.TotalAnalyses++

	if isAI {
		m.AIDetectedCount++
	}

	if isEarlyExit {
		m.EarlyExitCount++
	}

	// Calculate actual percentages
	if m.TotalAnalyses > 0 {
		m.AIDetectionRate = float64(m.AIDetectedCount) / float64(m.TotalAnalyses)
		m.EarlyExitRate = float64(m.EarlyExitCount) / float64(m.TotalAnalyses)
	}

	// Update cache hit rate
	totalCacheOps := m.CacheHits + m.CacheMisses
	if totalCacheOps > 0 {
		m.CacheHitRate = float64(m.CacheHits) / float64(totalCacheOps)
	}

	m.LastUpdate = time.Now()
}

func (m *Metrics) UpdateBusinessMetrics(result *analyzer.PipelineResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// ENTFERNT: m.TotalAnalyses++ - Das macht RecordVerdict schon

	// Track early exits from pipeline result
	if result.EarlyExit {
		m.EarlyExitCount++
	}

	// Track cache hits from pipeline result
	if result.CacheHit {
		m.CacheHits++
	} else {
		m.CacheMisses++
	}

	// Calculate percentages - basierend auf dem bereits gezählten Total
	if m.TotalAnalyses > 0 {
		m.EarlyExitRate = float64(m.EarlyExitCount) / float64(m.TotalAnalyses)
	}

	// Update cache hit rate
	totalCacheOps := m.CacheHits + m.CacheMisses
	if totalCacheOps > 0 {
		m.CacheHitRate = float64(m.CacheHits) / float64(totalCacheOps)
	}

	m.LastUpdate = time.Now()
}

func (m *Metrics) RecordVerdict(verdict string, isEarlyExit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalAnalyses++ // ← MUSS WIEDER AKTIVIERT WERDEN

	fmt.Printf("DEBUG RecordVerdict: verdict=%s, TotalAnalyses=%d\n", verdict, m.TotalAnalyses)

	// Determine if AI based on verdict string
	if verdict == "AI Generated (Confirmed)" ||
		verdict == "Very Likely AI Generated" ||
		verdict == "Likely AI Generated" ||
		verdict == "Possibly AI Generated" {
		m.AIDetectedCount++
		fmt.Printf("DEBUG: AI detected! Count now: %d\n", m.AIDetectedCount)
	}

	if isEarlyExit {
		m.EarlyExitCount++
	}

	// Calculate actual percentages
	if m.TotalAnalyses > 0 {
		m.AIDetectionRate = float64(m.AIDetectedCount) / float64(m.TotalAnalyses)
		m.EarlyExitRate = float64(m.EarlyExitCount) / float64(m.TotalAnalyses)
	}

	fmt.Printf("DEBUG: AI Detection Rate: %.3f (%d/%d)\n", m.AIDetectionRate, m.AIDetectedCount, m.TotalAnalyses)

	m.LastUpdate = time.Now()
}

// Cache tracking methods
func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheHits++

	// Recalculate cache hit rate
	totalCacheOps := m.CacheHits + m.CacheMisses
	if totalCacheOps > 0 {
		m.CacheHitRate = float64(m.CacheHits) / float64(totalCacheOps)
	}
}

func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheMisses++

	// Recalculate cache hit rate
	totalCacheOps := m.CacheHits + m.CacheMisses
	if totalCacheOps > 0 {
		m.CacheHitRate = float64(m.CacheHits) / float64(totalCacheOps)
	}
}

// Performance tracking methods
func (m *Metrics) RecordAnalysis(analysisType string, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Count analysis
	m.AnalysisCount[analysisType]++

	// Record duration
	if m.AnalysisDuration[analysisType] == nil {
		m.AnalysisDuration[analysisType] = make([]time.Duration, 0)
	}
	m.AnalysisDuration[analysisType] = append(m.AnalysisDuration[analysisType], duration)

	// Keep only last 100 durations to prevent memory growth
	if len(m.AnalysisDuration[analysisType]) > 100 {
		m.AnalysisDuration[analysisType] = m.AnalysisDuration[analysisType][1:]
	}

	// Record errors
	if err != nil {
		m.ErrorCount[analysisType]++
	}

	m.LastUpdate = time.Now()
}

func (m *Metrics) RecordSystemMetrics(memUsage, cpuUsage float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.MemoryUsage = append(m.MemoryUsage, memUsage)
	m.CPUUsage = append(m.CPUUsage, cpuUsage)

	// Keep only last 100 measurements
	if len(m.MemoryUsage) > 100 {
		m.MemoryUsage = m.MemoryUsage[1:]
	}
	if len(m.CPUUsage) > 100 {
		m.CPUUsage = m.CPUUsage[1:]
	}

	m.LastUpdate = time.Now()
}

func (m *Metrics) UpdateActiveConnections(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ActiveConnections = count
	m.LastUpdate = time.Now()
}

// Helper methods for calculations
func (m *Metrics) GetAverageDuration(analysisType string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	durations := m.AnalysisDuration[analysisType]
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func (m *Metrics) GetErrorRate(analysisType string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalCount := m.AnalysisCount[analysisType]
	if totalCount == 0 {
		return 0
	}

	errorCount := m.ErrorCount[analysisType]
	return float64(errorCount) / float64(totalCount)
}

func (m *Metrics) GetCurrentMemoryUsage() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.MemoryUsage) == 0 {
		return 0
	}
	return m.MemoryUsage[len(m.MemoryUsage)-1]
}

func (m *Metrics) GetCurrentCPUUsage() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.CPUUsage) == 0 {
		return 0
	}
	return m.CPUUsage[len(m.CPUUsage)-1]
}

// Export methods for API
func (m *Metrics) GetBusinessMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"ai_detection_rate": m.AIDetectionRate,
		"cache_hit_rate":    m.CacheHitRate,
		"early_exit_rate":   m.EarlyExitRate,
		"total_analyses":    m.TotalAnalyses,
		"ai_detected_count": m.AIDetectedCount,
		"early_exit_count":  m.EarlyExitCount,
	}
}

func (m *Metrics) GetOverallMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalAnalyses int64
	var totalErrors int64

	for _, count := range m.AnalysisCount {
		totalAnalyses += count
	}

	for _, errors := range m.ErrorCount {
		totalErrors += errors
	}

	var overallErrorRate float64
	if totalAnalyses > 0 {
		overallErrorRate = float64(totalErrors) / float64(totalAnalyses)
	}

	return map[string]interface{}{
		"total_analyses":     totalAnalyses,
		"total_errors":       totalErrors,
		"overall_error_rate": overallErrorRate,
	}
}

func (m *Metrics) GetPipelineMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDuration := m.GetAverageDuration("pipeline")
	errorRate := m.GetErrorRate("pipeline")
	totalCount := m.AnalysisCount["pipeline"]
	errorCount := m.ErrorCount["pipeline"]

	return map[string]interface{}{
		"average_duration": avgDuration.Milliseconds(),
		"error_rate":       errorRate,
		"total_count":      totalCount,
		"error_count":      errorCount,
	}
}

func (m *Metrics) GetUploadMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDuration := m.GetAverageDuration("upload")
	errorRate := m.GetErrorRate("upload")
	totalCount := m.AnalysisCount["upload"]
	errorCount := m.ErrorCount["upload"]

	return map[string]interface{}{
		"average_duration": avgDuration.Milliseconds(),
		"error_rate":       errorRate,
		"total_count":      totalCount,
		"error_count":      errorCount,
	}
}

func (m *Metrics) GetSystemMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"active_connections": m.ActiveConnections,
		"cache_hits":         m.CacheHits,
		"cache_misses":       m.CacheMisses,
		"memory_usage":       m.GetCurrentMemoryUsage(),
		"cpu_usage":          m.GetCurrentCPUUsage(),
	}
}

func (m *Metrics) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"business": map[string]interface{}{
			"total_analyses":    m.TotalAnalyses,
			"ai_detected_count": m.AIDetectedCount,
			"ai_detection_rate": m.AIDetectionRate,
			"early_exit_count":  m.EarlyExitCount,
			"early_exit_rate":   m.EarlyExitRate,
			"cache_hit_rate":    m.CacheHitRate,
		},
		"cache": map[string]interface{}{
			"hits":     m.CacheHits,
			"misses":   m.CacheMisses,
			"hit_rate": m.CacheHitRate,
		},
		"performance": map[string]interface{}{
			"analysis_count": m.AnalysisCount,
			"error_count":    m.ErrorCount,
		},
		"system": map[string]interface{}{
			"active_connections": m.ActiveConnections,
			"last_update":        m.LastUpdate.Unix(),
		},
	}
}

// Convenience method to get all metrics for API
func (m *Metrics) GetMetricsSummary() map[string]interface{} {
	return map[string]interface{}{
		"business": m.GetBusinessMetrics(),
		"overall":  m.GetOverallMetrics(),
		"pipeline": m.GetPipelineMetrics(),
		"upload":   m.GetUploadMetrics(),
		"system":   m.GetSystemMetrics(),
	}
}

// Reset metrics (useful for testing or periodic resets)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.AnalysisCount = make(map[string]int64)
	m.AnalysisDuration = make(map[string][]time.Duration)
	m.ErrorCount = make(map[string]int64)
	m.TotalAnalyses = 0
	m.AIDetectedCount = 0
	m.EarlyExitCount = 0
	m.AIDetectionRate = 0
	m.EarlyExitRate = 0
	m.CacheHitRate = 0
	m.CacheHits = 0
	m.CacheMisses = 0
	m.MemoryUsage = make([]float64, 0)
	m.CPUUsage = make([]float64, 0)
	m.ActiveConnections = 0
	m.LastUpdate = time.Now()
}

func (m *Metrics) IncrementActiveConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveConnections++
	m.LastUpdate = time.Now()
}

func (m *Metrics) DecrementActiveConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ActiveConnections > 0 {
		m.ActiveConnections--
	}
	m.LastUpdate = time.Now()
}

func (m *Metrics) RecordError(category string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ErrorCount == nil {
		m.ErrorCount = make(map[string]int64)
	}
	m.ErrorCount[category]++
	m.LastUpdate = time.Now()
}

// Success tracking methods - HINZUFÜGEN
func (m *Metrics) RecordSuccess(category string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.AnalysisCount == nil {
		m.AnalysisCount = make(map[string]int64)
	}
	m.AnalysisCount[category]++
	m.LastUpdate = time.Now()
}

// Duration tracking methods - HINZUFÜGEN
func (m *Metrics) RecordDuration(category string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.AnalysisDuration == nil {
		m.AnalysisDuration = make(map[string][]time.Duration)
	}

	// Keep only last 100 durations for memory efficiency
	durations := m.AnalysisDuration[category]
	if len(durations) >= 100 {
		durations = durations[1:]
	}
	durations = append(durations, duration)
	m.AnalysisDuration[category] = durations

	m.LastUpdate = time.Now()
}
