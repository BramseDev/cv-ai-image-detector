package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GenerateObjectCoherenceExplanation(data map[string]interface{}) string {
	if aiProbability, exists := utils.GetFloatValue(data, "ai_probability"); exists {
		if aiProbability >= 0.7 {
			return fmt.Sprintf("Object coherence analysis finds inconsistent object relationships (%.1f%% AI indicators). Objects don't follow natural physics or perspective.", aiProbability*100)
		} else if aiProbability >= 0.5 {
			return fmt.Sprintf("Some object inconsistencies detected (%.1f%% AI indicators). Mixed coherence signals in object placement.", aiProbability*100)
		} else if aiProbability >= 0.3 {
			return fmt.Sprintf("Object relationships appear mostly consistent (%.1f%% AI indicators). Natural object placement patterns.", aiProbability*100)
		} else {
			return fmt.Sprintf("Object coherence analysis shows natural relationships (%.1f%% AI indicators). Consistent with real-world object physics.", aiProbability*100)
		}
	}

	return "Object coherence analysis completed. Spatial relationships evaluated for natural consistency."
}
