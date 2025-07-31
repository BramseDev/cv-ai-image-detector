package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GeneratePixelExplanation(data map[string]interface{}) string {
	if aiProbability, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		if aiProbability >= 0.7 {
			return fmt.Sprintf("Pixel-level analysis reveals mathematical irregularities (%.1f%% AI indicators). Statistical patterns inconsistent with natural photography.", aiProbability*100)
		} else if aiProbability >= 0.5 {
			return fmt.Sprintf("Pixel analysis shows some mathematical anomalies (%.1f%% AI indicators). Mixed statistical signatures detected.", aiProbability*100)
		} else if aiProbability >= 0.3 {
			return fmt.Sprintf("Pixel patterns appear mostly natural (%.1f%% AI indicators). Standard noise and grain characteristics.", aiProbability*100)
		} else {
			return fmt.Sprintf("Pixel-level analysis shows natural characteristics (%.1f%% AI indicators). Consistent with authentic camera sensor data.", aiProbability*100)
		}
	}

	return "Pixel-level analysis completed. Mathematical patterns analyzed for authenticity markers."
}
