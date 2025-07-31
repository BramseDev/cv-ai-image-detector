package analysis

import (
	"fmt"
	"log/slog"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/explanations"
	exifanalyzer "github.com/BramseDev/imageAnalyzer/pkg/analyzer/exif_analyzer"

	"github.com/BramseDev/imageAnalyzer/pkg/analyzer/pipeline"
)

func CreateStructuredResponse(results *pipeline.PipelineResult) map[string]interface{} {
	response := make(map[string]interface{})

	// Debug output
	fmt.Println("\n=== PIPELINE DEBUG ===")
	fmt.Printf("Stages executed: %d\n", len(results.Results))
	fmt.Printf("Results available: %d\n", len(results.Results))

	fmt.Println("\nStages run:")
	for _, stage := range results.StagesRun {
		fmt.Printf("  - %s\n", stage)
	}

	fmt.Println("\nResults received:")
	for name, result := range results.Results {
		fmt.Printf("  - %s: %T\n", name, result)
	}
	fmt.Println("======================\n")

	// Convert results
	convertedResults := make(map[string]map[string]interface{})
	fmt.Println("\n=== CONVERTED RESULTS ===")
	for name, result := range results.Results {
		converted, err := convertAnalysisResult(result)
		if err != nil {
			slog.Error("Failed to convert result", "stage", name, "error", err)
			fmt.Printf("  - %s: err='%v'\n", name, err)
			continue
		}
		convertedResults[name] = converted
		fmt.Printf("  - %s: err=''\n", name)
	}
	fmt.Println("=========================\n")

	// Create analysis sections
	for name, data := range convertedResults {
		if section := createAnalysisSection(name, data); section != nil {
			response[name] = section
		}
	}

	return response
}

func convertAnalysisResult(data interface{}) (map[string]interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return v, nil
	case *exifanalyzer.EXIFData:
		result := make(map[string]interface{})
		if v != nil {
			// Verwende die tatsächlichen Feldnamen aus EXIFData
			hasCameraInfo := (v.Make != "" || v.Model != "")
			result["has_camera_info"] = hasCameraInfo
			result["camera_make"] = v.Make
			result["camera_model"] = v.Model

			if v.DateTime != nil {
				result["creation_date"] = v.DateTime.Format("2006-01-02 15:04:05")
			}

			if v.GPS != nil {
				result["gps_info"] = map[string]float64{
					"latitude":  v.GPS[0],
					"longitude": v.GPS[1],
				}
			}

			// Raw EXIF als Indikator für Modifikation
			result["modification_indicators"] = len(v.Raw) > 0
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported result type: %T", data)
	}
}

func createAnalysisSection(analysisType string, data map[string]interface{}) map[string]interface{} {
	section := map[string]interface{}{
		"analysis_type": analysisType,
		"data":          data,
	}

	// Add type-specific explanations
	switch analysisType {
	case "exif":
		section["explanation"] = explanations.GenerateEXIFExplanation(data)
	case "ai-model":
		section["explanation"] = explanations.GenerateAIExplanation(data)
	case "metadata":
		section["explanation"] = explanations.GenerateMetadataExplanation(data)
	case "compression":
		section["explanation"] = explanations.GenerateCompressionExplanation(data)
	case "pixel-analysis":
		section["explanation"] = explanations.GeneratePixelExplanation(data)
	case "lighting-analysis":
		section["explanation"] = explanations.GenerateLightingExplanation(data)
	case "color-balance":
		section["explanation"] = explanations.GenerateColorExplanation(data)
	case "artifacts":
		section["explanation"] = explanations.GenerateArtifactsExplanation(data)
	case "advanced-artifacts":
		section["explanation"] = explanations.GenerateAdvancedArtifactsExplanation(data)
	case "object-coherence":
		section["explanation"] = explanations.GenerateObjectCoherenceExplanation(data)
	case "metadata-quick":
		section["explanation"] = explanations.GenerateQuickMetadataExplanation(data)
	default:
		section["explanation"] = explanations.GenerateTraditionalExplanation(data)
	}

	return section
}
