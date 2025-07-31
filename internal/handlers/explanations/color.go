package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GenerateColorExplanation(data map[string]interface{}) string {
	if aiProbability, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		if aiProbability >= 0.7 {
			return fmt.Sprintf("Color balance analysis shows artificial color distribution (%.1f%% AI indicators). Unnatural color relationships suggest algorithmic generation.", aiProbability*100)
		} else if aiProbability >= 0.5 {
			return fmt.Sprintf("Color patterns show some artificial characteristics (%.1f%% AI indicators). Mixed color balance signals.", aiProbability*100)
		} else if aiProbability >= 0.3 {
			return fmt.Sprintf("Color balance appears mostly natural (%.1f%% AI indicators). Standard color distribution patterns.", aiProbability*100)
		} else {
			return fmt.Sprintf("Color analysis shows natural characteristics (%.1f%% AI indicators). Consistent with authentic color capture.", aiProbability*100)
		}
	}

	return "Color balance analysis completed. Color relationships evaluated for natural distribution."
}
