package verdict

import "math"

func determineBalancedVerdict(score float64, scores map[string]float64) (string, float64) {

	highConfidenceScores := 0
	lowConfidenceScores := 0

	for _, individualScore := range scores {
		if individualScore >= 0.7 {
			highConfidenceScores++
		} else if individualScore <= 0.3 {
			lowConfidenceScores++
		}
	}
	totalScores := len(scores)
	baseConfidence := 0.6

	if totalScores > 0 {
		definitiveRatio := float64(highConfidenceScores+lowConfidenceScores) / float64(totalScores)
		baseConfidence = 0.5 + (definitiveRatio * 0.4)
	}

	thresholdAdjustment := 0.0
	if highConfidenceScores >= 3 {
		thresholdAdjustment = -0.05
	} else if lowConfidenceScores >= 3 {
		thresholdAdjustment = 0.05
	}

	scorePercent := score * 100

	// Verdict-Bestimmung mit neuen Schwellenwerten
	if scorePercent >= (80.0 + thresholdAdjustment*100) {
		confidence := math.Min(0.95, baseConfidence+0.25)
		return "Very Likely AI Generated", confidence
	} else if scorePercent >= (60.0 + thresholdAdjustment*100) {
		confidence := math.Min(0.9, baseConfidence+0.15)
		return "Likely AI Generated", confidence
	} else if scorePercent >= (55.0 + thresholdAdjustment*100) {
		confidence := math.Min(0.8, baseConfidence)
		return "Possibly AI Generated", confidence
	} else if scorePercent >= (20.0 + thresholdAdjustment*100) {
		confidence := math.Min(0.85, baseConfidence+0.1)
		return "Likely Authentic", confidence
	} else {
		confidence := math.Min(0.9, baseConfidence+0.2)
		return "Very Likely Authentic", confidence
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
	avgScore := sum / float64(len(scores))

	var variance float64
	for _, score := range scores {
		variance += (score - avgScore) * (score - avgScore)
	}
	variance /= float64(len(scores))

	consistency := 1.0 - variance
	if consistency < 0.0 {
		consistency = 0.0
	}
	if consistency > 1.0 {
		consistency = 1.0
	}

	return consistency
}
