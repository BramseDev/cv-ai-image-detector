package verdict

import "math"

func applyBalancedCalibration(scores map[string]float64) map[string]float64 {
	calibratedScores := make(map[string]float64)

	calibrationFactors := map[string]float64{
		"lighting-analysis":  1.30,
		"artifacts":          1.05,
		"advanced-artifacts": 1.15,
		"pixel-analysis":     1.05,
		"color-balance":      1.00,
		"object-coherence":   0.90,
		"compression":        0.40,
		"metadata":           1.0,
		"c2pa":               1.0,
		"exif":               0.70,
		"ai-model":           1.0,
		"metadata-quick":     1.0,
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

	// Dynamic adjustments based on scores
	if metadataScore, exists := scores["metadata"]; exists && metadataScore <= 0.1 {
		adjustedWeights["metadata"] *= 1.3 // Boost metadata weight if strong authenticity
	}

	if exifScore, exists := scores["exif"]; exists && exifScore <= 0.1 {
		adjustedWeights["exif"] *= 1.3 // Boost EXIF weight if strong authenticity
	}

	if compressionScore, exists := scores["compression"]; exists && compressionScore >= 0.6 {
		adjustedWeights["compression"] *= 1.3 // Boost compression weight if suspicious
	}

	if pixelScore, exists := scores["pixel-analysis"]; exists && pixelScore >= 0.6 {
		adjustedWeights["pixel-analysis"] *= 1.3 // Boost pixel analysis weight if suspicious
	}

	if colorScore, exists := scores["color-balance"]; exists && colorScore <= 0.2 {
		adjustedWeights["color-balance"] *= 1.3 // Boost color balance weight if very authentic
	}

	return adjustedWeights
}
