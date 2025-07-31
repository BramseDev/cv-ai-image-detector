package verdict

func calculateAdvancedBoost(scores map[string]float64) float64 {
	boost := 1.0

	consistencyCount := 0
	traditionalMethods := []string{"artifacts", "lighting-analysis", "advanced-artifacts", "pixel-analysis"}
	for _, method := range traditionalMethods {
		if score, exists := scores[method]; exists {
			if score >= 0.4 {
				consistencyCount++
			}
		}
	}

	if consistencyCount >= 3 {
		boost *= 1.4
	} else if consistencyCount >= 2 {
		boost *= 1.2
	}

	if aiModelScore, exists := scores["ai-model"]; exists {
		if artifactsScore, exists := scores["artifacts"]; exists {
			if aiModelScore >= 0.5 && artifactsScore >= 0.6 {
				boost *= 1.5
			}

			if aiModelScore >= 0.7 && artifactsScore <= 0.3 {
				boost *= 1.3
			}
		}

		if aiModelScore <= 0.3 {
			authenticityCount := 0
			if compressionScore, exists := scores["compression"]; exists && compressionScore <= 0.3 {
				authenticityCount++
			}
			if colorScore, exists := scores["color-balance"]; exists && colorScore <= 0.2 {
				authenticityCount++
			}
			if artifactsScore, exists := scores["artifacts"]; exists && artifactsScore <= 0.4 {
				authenticityCount++
			}

			if authenticityCount >= 2 {
				boost *= 0.6
			}
		}

		if aiModelScore >= 0.8 || aiModelScore <= 0.2 {
			boost *= 1.4
		}
	}

	if boost > 2.0 {
		boost = 2.0
	} else if boost < 0.4 {
		boost = 0.4
	}

	return boost
}
