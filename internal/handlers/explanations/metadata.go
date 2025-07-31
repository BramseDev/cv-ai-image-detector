package explanations

import "github.com/BramseDev/imageAnalyzer/internal/handlers/utils"

func GenerateMetadataExplanation(data map[string]interface{}) string {
	if hasMetadata, exists := data["has_metadata"]; exists {
		if has, ok := hasMetadata.(bool); ok && has {
			// Check for specific metadata fields
			if software, exists := utils.GetStringValue(data, "software"); exists && software != "" {
				return "Rich metadata found including software information. Strong authenticity indicator from real camera/software."
			}
			return "Comprehensive metadata present. Indicates authentic photograph with proper digital signature."
		}
	}

	return "Limited or missing metadata. Could indicate AI generation or heavy post-processing that stripped metadata."
}

func GenerateQuickMetadataExplanation(data map[string]interface{}) string {
	if hasMetadata, exists := data["has_metadata"]; exists {
		if has, ok := hasMetadata.(bool); ok && has {
			return "Basic metadata scan shows positive indicators for authentic image."
		}
	}

	return "Quick metadata scan shows minimal metadata present."
}
