package verdict

import (
	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func calculateEXIFScore(data map[string]interface{}) float64 {
	hasCameraInfo, _ := utils.GetFloatValue(data, "has_camera_info")
	if hasCameraInfo > 0 {
		return 0.0
	}
	return 0.0
}

func calculateMetadataScore(data map[string]interface{}) float64 {
	if hasMetadata, exists := data["has_metadata"]; exists {
		if has, ok := hasMetadata.(bool); ok && has {
			return 0.0
		}
	}
	return 0.0
}

func calculateAIModelScore(data map[string]interface{}) float64 {
	if probability, exists := utils.GetFloatValue(data, "probability"); exists {
		return probability
	}
	return 0.5
}

func calculateArtifactsScore(data map[string]interface{}) float64 {
	if score, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		return score
	}
	return 0.4
}

func calculateCompressionScore(data map[string]interface{}) float64 {
	if score, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		return score
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

func calculateAdvancedArtifactsScore(data map[string]interface{}) float64 {
	if score, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		return score
	}
	return 0.333
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
