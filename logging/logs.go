package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

type Logger struct {
	*slog.Logger
}

func NewLogger(level slog.Level) *Logger {
	os.MkdirAll("logs", 0755)

	logFile, err := os.OpenFile(
		fmt.Sprintf("logs/app-%s.log", time.Now().Format("2006-01-02")),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		panic(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

func (l *Logger) LogAnalysisStart(imagePath string, analysisType string) {
	l.Info("Analysis started",
		"type", analysisType,
		"image", imagePath,
		"timestamp", time.Now().Unix(),
	)
}

func (l *Logger) LogAnalysisComplete(analysisType string, duration time.Duration, success bool) {
	l.Info("Analysis completed",
		"type", analysisType,
		"duration_ms", duration.Milliseconds(),
		"success", success,
	)
}

func (l *Logger) LogPipelineMetrics(stagesRun []string, totalDuration time.Duration, earlyExit bool) {
	l.Info("Pipeline metrics",
		"stages_completed", len(stagesRun),
		"stages", stagesRun,
		"total_duration_ms", totalDuration.Milliseconds(),
		"early_exit", earlyExit,
	)
}
