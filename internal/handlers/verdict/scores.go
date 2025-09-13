package verdict

import (
	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func calculateEXIFScore(data map[string]interface{}) float64 {
	if hasCameraInfo, exists := data["has_camera_info"]; exists {
		if has, ok := hasCameraInfo.(bool); ok && has {
			return 0.0 // Authentisch - Camera-Info vorhanden
		} else {
			return -1 // Ignorieren - keine Camera-Info
		}
	}
	return -1 // Ignorieren - kein EXIF-Check möglich
}

func calculateMetadataScore(data map[string]interface{}) float64 {
	if hasMetadata, exists := data["has_metadata"]; exists {
		if has, ok := hasMetadata.(bool); ok && has {
			return 0.0 // Authentisch - Metadata vorhanden
		} else {
			return -1 // Ignorieren - keine Metadata
		}
	}
	return -1 // Ignorieren - kein Metadata-Check möglich
}

func calculateAIModelScore(data map[string]interface{}) float64 {
	if probability, exists := utils.GetFloatValue(data, "probability"); exists {
		return probability
	}
	return 0.5
}

func calculatePixelScore(data map[string]interface{}) float64 {
	if score, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		return score
	}
	return 0.0
}

func calculateLightingScore(data map[string]interface{}) float64 {
	if score, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		return score
	}
	return 0.667
}

func calculateColorScore(data map[string]interface{}) float64 {
	if score, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		return score
	}
	return 0.5
}

func calculateObjectCoherenceScore(data map[string]interface{}) float64 {
	if score, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		return score
	}
	return 0.333
}

func calculateQuickMetadataScore(data map[string]interface{}) float64 {
	if hasMetadata, exists := data["has_metadata"]; exists {
		if has, ok := hasMetadata.(bool); ok && has {
			return 0.0
		}
	}
	return 0.0
}
func calculateArtifactsScore(data map[string]interface{}) float64 {
	// Check für overall_assessment direkt (deine Daten!)
	if score, exists := utils.GetFloatValue(data, "overall_assessment"); exists {
		return score
	}

	// Fallback für alte Datenstruktur
	if assessment, exists := data["overall_assessment"].(map[string]interface{}); exists {
		if score, exists := utils.GetFloatValue(assessment, "ai_probability_score"); exists {
			return score
		}
	}

	return 0.4 // Default
}

func calculateAdvancedArtifactsScore(data map[string]interface{}) float64 {
	// ERSTE PRIORITÄT: advanced_assessment verwenden
	if score, exists := utils.GetFloatValue(data, "advanced_assessment"); exists {
		return score
	}

	// ZWEITE PRIORITÄT: advanced_ai_probability
	if advancedData, exists := data["advanced_assessment"].(map[string]interface{}); exists {
		if score, exists := utils.GetFloatValue(advancedData, "advanced_ai_probability"); exists {
			return score
		}
	}

	// FALLBACK: alte Struktur
	if score, exists := utils.GetFloatValue(data, "advanced_ai_probability"); exists {
		return score
	}

	return 0.333 // Default
}

func calculateCompressionScore(data map[string]interface{}) float64 {
	// Iteriere durch alle Dateien im compression data
	for _, fileData := range data {
		if fileMap, ok := fileData.(map[string]interface{}); ok {
			if analysis, exists := fileMap["compression_ai_analysis"]; exists {
				if analysisMap, ok := analysis.(map[string]interface{}); ok {
					if score, exists := utils.GetFloatValue(analysisMap, "ai_probability"); exists {
						return score
					}
				}
			}
		}
	}
	return 0.5 // Default
}

func calculatePixelAnalysisScore(data map[string]interface{}) float64 {
	if assessment, exists := data["overall_assessment"].(map[string]interface{}); exists {
		if score, exists := utils.GetFloatValue(assessment, "ai_probability_score"); exists {
			return score
		}
	}
	return 0.5 // Default
}

func calculateLightingAnalysisScore(data map[string]interface{}) float64 {
	if lighting, exists := data["lighting_analysis"].(map[string]interface{}); exists {
		if score, exists := utils.GetFloatValue(lighting, "ai_lighting_score"); exists {
			return score
		}
	}
	return 0.5 // Default
}

func calculateColorBalanceScore(data map[string]interface{}) float64 {
	if score, exists := utils.GetFloatValue(data, "ai_color_score"); exists {
		return score
	}
	return 0.5 // Default
}

func calculateMetadataQuickScore(data map[string]interface{}) float64 {
	// Metadata-Quick hat normalerweise keinen AI-Score
	return -1 // Ignorieren
}

func calculateC2PAScore(data map[string]interface{}) float64 {
	// Prüfe ob überhaupt Claims gefunden wurden
	if claimsFoundValue, exists := data["claims_found"]; exists {
		if claimsFound, ok := claimsFoundValue.(bool); ok && !claimsFound {
			return -1 // Ignorieren - keine C2PA-Daten vorhanden
		}
	}

	// Prüfe Claims Count
	if claimsCount, exists := utils.GetFloatValue(data, "claims_count"); exists {
		if claimsCount == 0 {
			return -1 // Ignorieren - keine Claims
		}
	}

	// Nur wenn tatsächlich Claims vorhanden sind, verwende den Score
	if score, exists := utils.GetFloatValue(data, "score"); exists {
		return score / 100.0 // Normalisiere auf 0-1
	}

	return -1 // Default: Ignorieren
}
