package verdict

import "math"

func applyBalancedCalibration(scores map[string]float64) map[string]float64 {
	calibratedScores := make(map[string]float64)

	calibrationFactors := map[string]float64{
		// Boost the working methods even more
		"compression":        1.0,
		"artifacts":          1.0,
		"pixel-analysis":     1.4,
		"lighting-analysis":  1.5,
		"color-balance":      1.4,
		"advanced-artifacts": 1.0,

		"c2pa":           0.9,
		"exif":           0.85,
		"metadata":       1.0,
		"metadata-quick": 0.8,

		"object-coherence": 0.8,

		"ai-model": 1.2,
	}

	for name, score := range scores {
		if factor, exists := calibrationFactors[name]; exists {
			calibratedScore := score * factor
			calibratedScores[name] = math.Min(1.0, calibratedScore)
		} else {
			calibratedScores[name] = score // FALLBACK für fehlende Faktoren
		}
	}

	return calibratedScores
}

func applyDynamicWeights(weights map[string]float64, scores map[string]float64) map[string]float64 {
	adjustedWeights := make(map[string]float64)

	// Copy base weights
	for method, weight := range weights {
		adjustedWeights[method] = weight
	}

	if exifScore, exists := scores["exif"]; exists {
		if exifScore >= 0.8 {
			adjustedWeights["exif"] *= 1.4
		} else if exifScore <= 0.2 {
			adjustedWeights["exif"] *= 1.3
		}
	}

	if colorScore, exists := scores["color-balance"]; exists {
		if colorScore >= 0.7 {
			adjustedWeights["color-balance"] *= 1.3
		} else if colorScore <= 0.3 {
			adjustedWeights["color-balance"] *= 1.2
		}
	}

	if lightingScore, exists := scores["lighting-analysis"]; exists && lightingScore >= 0.6 {
		adjustedWeights["lighting-analysis"] *= 1.3
	}

	if compressionScore, exists := scores["compression"]; exists && compressionScore >= 0.4 {
		adjustedWeights["compression"] *= 0.5
	}

	if aiScore, exists := scores["ai-model"]; exists && aiScore >= 0 {
		adjustedWeights["ai-model"] *= 1.5
	}

	return adjustedWeights
}
