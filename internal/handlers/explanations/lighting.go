package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GenerateLightingExplanation(data map[string]interface{}) string {
	if aiProbability, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		if aiProbability >= 0.7 {
			return fmt.Sprintf("Lighting analysis reveals inconsistent light sources (%.1f%% AI indicators). Shadows and highlights don't follow physical light laws.", aiProbability*100)
		} else if aiProbability >= 0.5 {
			return fmt.Sprintf("Lighting patterns show some inconsistencies (%.1f%% AI indicators). Mixed lighting characteristics detected.", aiProbability*100)
		} else if aiProbability >= 0.3 {
			return fmt.Sprintf("Lighting appears mostly consistent (%.1f%% AI indicators). Natural light distribution patterns.", aiProbability*100)
		} else {
			return fmt.Sprintf("Lighting analysis shows natural characteristics (%.1f%% AI indicators). Consistent with real-world lighting conditions.", aiProbability*100)
		}
	}

	return "Lighting analysis completed. Physical light consistency evaluated for authenticity."
}
