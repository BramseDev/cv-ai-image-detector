package verdict

func calculateAdvancedBoost(scores map[string]float64) float64 {
	boost := 1.0

	topMethods := []string{"color-balance", "lighting-analysis", "metadata"}
	aiConsistency := 0
	authenticityConsistency := 0
	
	for _, method := range topMethods {
		if score, exists := scores[method]; exists {
			if score >= 0.6 {
				aiConsistency++
			} else if score <= 0.4 {
				authenticityConsistency++
			}
		}
	}

	if aiConsistency >= 2 {
		boost *= 1.25
	} else if authenticityConsistency >= 2 {
		boost *= 0.75
	}

	if aiModelScore, exists := scores["ai-model"]; exists {
		if aiModelScore >= 0.8 {
			boost *= 1.4
		} else if aiModelScore <= 0.2 {
			boost *= 0.65
		}
		
		if colorScore, exists := scores["color-balance"]; exists {
			if aiModelScore >= 0.7 && colorScore >= 0.6 {
				boost *= 1.3
			} else if aiModelScore <= 0.3 && colorScore <= 0.4 {
				boost *= 0.7
			}
		}
	}

	if colorScore, exists := scores["color-balance"]; exists {
		if metadataScore, exists := scores["metadata"]; exists {
			if colorScore >= 0.6 && metadataScore >= 0.6 {
				boost *= 1.2
			} else if colorScore <= 0.4 && metadataScore <= 0.4 {
				boost *= 0.8
			}
		}
	}

	if compressionScore, exists := scores["compression"]; exists && compressionScore >= 0.5 {
		boost *= 0.9
	}

	if exifScore, exists := scores["exif"]; exists && exifScore >= 0.8 {
		if colorScore, exists := scores["color-balance"]; exists && colorScore <= 0.4 {
			boost *= 0.85
		}
	}

	if boost > 1.6 {
		boost = 1.6
	} else if boost < 0.6 {
		boost = 0.6
	}

	return boost
}
