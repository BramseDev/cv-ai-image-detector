package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GenerateCompressionExplanation(data map[string]interface{}) string {
	if aiProbability, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		if aiProbability >= 0.7 {
			return fmt.Sprintf("Compression analysis shows artificial patterns (%.1f%% AI indicators). Unusual compression artifacts suggest algorithmic generation.", aiProbability*100)
		} else if aiProbability >= 0.5 {
			return fmt.Sprintf("Compression patterns show some artificial characteristics (%.1f%% AI indicators). Mixed compression signals detected.", aiProbability*100)
		} else if aiProbability >= 0.3 {
			return fmt.Sprintf("Compression analysis shows mostly natural patterns (%.1f%% AI indicators). Standard JPEG compression detected.", aiProbability*100)
		} else {
			return fmt.Sprintf("Compression patterns appear highly natural (%.1f%% AI indicators). Consistent with authentic camera compression.", aiProbability*100)
		}
	}

	return "Compression analysis completed. Patterns consistent with standard image processing."
}
