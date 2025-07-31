package analysis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/BramseDev/imageAnalyzer/pkg/analyzer/pipeline"
)

func RunSecureAnalyses(tempFile string, logger *slog.Logger) (*pipeline.PipelineResult, error) {

	analysisPipeline := pipeline.NewAnalysisPipeline()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	results, err := analysisPipeline.RunAnalysis(ctx, tempFile)
	if err != nil {
		return nil, fmt.Errorf("pipeline failed: %w", err)
	}

	logger.Info("Pipeline completed successfully",
		"duration", results.ProcessTime.Nanoseconds(),
		"stages_run", len(results.StagesRun),
		"confidence", results.Confidence,
		"cache_hit", results.CacheHit,
	)

	return results, nil
}
