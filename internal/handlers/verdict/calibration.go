package verdict

import "math"

func applyBalancedCalibration(scores map[string]float64) map[string]float64 {
	calibratedScores := make(map[string]float64)

	calibrationFactors := map[string]float64{
		"lighting-analysis":  1.30,
		"artifacts":          1.05,
		"advanced-artifacts": 0.95,
		"pixel-analysis":     0.80,
		"color-balance":      1.20,
		"object-coherence":   1.10,
		"compression":        0.20,
		"metadata":           1.15,
		"c2pa":               0.90,
		"exif":               0.70,
		"ai-model":           1.25,
		"metadata-quick":     1.08,
	}

	for name, score := range scores {
		if factor, exists := calibrationFactors[name]; exists {
			calibratedScore := score * factor
			calibratedScores[name] = math.Min(1.0, calibratedScore)
		} else {
			calibratedScores[name] = score
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
