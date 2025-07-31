package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GenerateArtifactsExplanation(data map[string]interface{}) string {
	if aiProbability, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		if aiProbability >= 0.7 {
			return fmt.Sprintf("Artifact analysis detects strong AI generation markers (%.1f%% AI indicators). Multiple artificial patterns found.", aiProbability*100)
		} else if aiProbability >= 0.5 {
			return fmt.Sprintf("Some AI artifacts detected (%.1f%% AI indicators). Moderate artificial pattern indicators.", aiProbability*100)
		} else if aiProbability >= 0.3 {
			return fmt.Sprintf("Minimal artificial artifacts found (%.1f%% AI indicators). Mostly natural image characteristics.", aiProbability*100)
		} else {
			return fmt.Sprintf("Very few AI artifacts detected (%.1f%% AI indicators). Strong indicators of authentic photography.", aiProbability*100)
		}
	}

	return "Artifact analysis completed. Image examined for AI generation signatures."
}

func GenerateAdvancedArtifactsExplanation(data map[string]interface{}) string {
	if aiProbability, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		if aiProbability >= 0.7 {
			return fmt.Sprintf("Advanced artifact detection found sophisticated AI patterns (%.1f%% AI indicators). Deep analysis reveals algorithmic fingerprints.", aiProbability*100)
		} else if aiProbability >= 0.5 {
			return fmt.Sprintf("Advanced analysis shows some AI characteristics (%.1f%% AI indicators). Subtle artificial patterns detected.", aiProbability*100)
		} else if aiProbability >= 0.3 {
			return fmt.Sprintf("Advanced artifact scan shows mostly authentic patterns (%.1f%% AI indicators). Natural image structure preserved.", aiProbability*100)
		} else {
			return fmt.Sprintf("Advanced analysis confirms authentic characteristics (%.1f%% AI indicators). No sophisticated AI artifacts found.", aiProbability*100)
		}
	}

	return "Advanced artifact analysis completed. Deep pattern recognition applied for AI detection."
}
