package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GenerateTraditionalExplanation(data map[string]interface{}) string {
	if aiProbability, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		if aiProbability >= 0.7 {
			return fmt.Sprintf("Computer Vision detected AI patterns in multiple analysis areas (%.1f%% AI probability). Strong algorithmic signatures found.", aiProbability*100)
		} else if aiProbability >= 0.5 {
			return fmt.Sprintf("Computer Vision shows mixed AI indicators (%.1f%% AI probability). Some artificial characteristics detected.", aiProbability*100)
		} else if aiProbability >= 0.3 {
			return fmt.Sprintf("Computer Vision analysis shows mostly authentic patterns (%.1f%% authenticity). Natural image characteristics dominate.", (1-aiProbability)*100)
		} else {
			return fmt.Sprintf("Computer Vision strongly indicates authentic photography (%.1f%% authenticity). Natural patterns throughout analysis.", (1-aiProbability)*100)
		}
	}

	return "Computer Vision analysis completed with standard pattern recognition techniques."
}
