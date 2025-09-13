package verdict

import "math"

func calculateAdvancedBoost(scores map[string]float64) float64 {
	boost := 1.0

	// AI-Model Vertrauen einbeziehen
	aiModelScore := scores["ai-model"]

	if aiModelScore < 0.7 {
		boost *= 0.8 // Reduziere boost bei unsicherem AI-Model
	}

	traditionalMethods := []string{"color-balance", "lighting-analysis", "artifacts", "advanced-artifacts", "pixel-analysis", "object-coherence"}
	aiConsistency := 0

	for _, method := range traditionalMethods {
		if score, exists := scores[method]; exists {
			if score >= 0.6 { // Höhere Schwelle: War 0.45, jetzt 0.6
				aiConsistency++
			}
		}
	}

	if aiConsistency >= 4 { // War 3, jetzt 4
		boost *= 1.4 // War 1.8, jetzt 1.4
	} else if aiConsistency >= 3 { // War 2, jetzt 3
		boost *= 1.2 // War 1.5, jetzt 1.2
	}

	return math.Min(boost, 1.8) // War 2.2, jetzt 1.8
}
