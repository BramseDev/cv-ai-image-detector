package verdict

import "math"

func determineBalancedVerdict(score float64, scores map[string]float64) (string, float64) {
	baseConfidence := calculateConfidence(scores, len(scores))

	thresholdAdjustment := 0.0
	if len(scores) < 4 {
		thresholdAdjustment = 0.05
	}

	if score >= (0.75 + thresholdAdjustment) {
		confidence := math.Min(0.95, baseConfidence+0.15)
		return "AI Generated", confidence
	} else if score >= (0.60 + thresholdAdjustment) {
		confidence := math.Min(0.85, baseConfidence+0.10)
		return "Likely AI Generated", confidence
	} else if score >= (0.59 + thresholdAdjustment) {
		confidence := math.Min(0.85, baseConfidence+0.10)
		return "Likely Authentic", confidence
	} else {
		confidence := math.Min(0.40, baseConfidence+0.15)
		return "Authentic", confidence
	}
}

func calculateConfidence(scores map[string]float64, totalMethods int) float64 {
	if totalMethods == 0 {
		return 0.5
	}

	methodConfidence := float64(len(scores)) / float64(totalMethods)

	consistencyBonus := calculateConsistency(scores)

	confidence := (methodConfidence * 0.7) + (consistencyBonus * 0.3)

	if confidence < 0.5 {
		confidence = 0.5
	}
	if confidence > 0.99 {
		confidence = 0.99
	}

	return confidence
}

func calculateConsistency(scores map[string]float64) float64 {
	if len(scores) < 2 {
		return 0.5
	}

	var sum float64
	for _, score := range scores {
		sum += score
	}
	mean := sum / float64(len(scores))

	var variance float64
	for _, score := range scores {
		diff := score - mean
		variance += diff * diff
	}
	variance /= float64(len(scores))

	consistency := math.Max(0, 1.0-variance)
	return consistency
}
