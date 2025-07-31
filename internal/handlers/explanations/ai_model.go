package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GenerateAIExplanation(data map[string]interface{}) string {
	if probability, exists := utils.GetFloatValue(data, "probability"); exists {
		if probability >= 0.8 {
			return fmt.Sprintf("Neural network strongly predicts AI-generated (%.1f%% confidence). Very high certainty from deep learning model.", probability*100)
		} else if probability >= 0.6 {
			return fmt.Sprintf("Neural network indicates likely AI-generated (%.1f%% confidence). Moderate to high certainty.", probability*100)
		} else if probability >= 0.4 {
			return fmt.Sprintf("Neural network shows mixed signals (%.1f%% confidence). Uncertain classification.", probability*100)
		} else if probability >= 0.2 {
			return fmt.Sprintf("Neural network suggests likely authentic (%.1f%% confidence). Moderate authenticity indicators.", probability*100)
		} else {
			return fmt.Sprintf("Neural network strongly suggests authentic (%.1f%% confidence). High confidence in real photography.", probability*100)
		}
	}

	return "AI model analysis completed with standard confidence level."
}
