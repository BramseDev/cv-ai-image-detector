package verdict

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/pkg/analyzer/pipeline"
)

func CalculateOverallVerdict(results *pipeline.PipelineResult) map[string]interface{} {
	scores := make(map[string]float64)
	var reasoning []string

	tempPipeline := pipeline.NewAnalysisPipeline()

	traditionalMethods := []string{
		"artifacts", "compression", "pixel-analysis", "color-balance",
		"lighting-analysis", "advanced-artifacts", "object-coherence",
	}

	metadataMethods := []string{
		"metadata", "metadata-quick", "exif", "c2pa",
	}

	aiMethods := []string{
		"ai-model",
	}

	// Separate Score-Sammlung
	traditionalScores := make(map[string]float64)
	metadataScores := make(map[string]float64)
	aiScores := make(map[string]float64)

	// DEBUG: Log alle rohen Scores
	fmt.Printf("\n=== DEBUG SCORES ===\n")

	weights := map[string]float64{
		"metadata":           3.5,
		"c2pa":               3.5,
		"artifacts":          2.2,
		"lighting-analysis":  2.6,
		"advanced-artifacts": 2.0,
		"pixel-analysis":     1.8,
		"color-balance":      1.5,
		"object-coherence":   1.2,
		"compression":        2.5,
		"exif":               2.8,
		"ai-model":           4.0,
		"metadata-quick":     1.0,
	}

	var definitiveScore float64 = -1

	for name, result := range results.Results {
		score := tempPipeline.ExtractConfidenceScore(result)
		if score >= 0 {
			scores[name] = score

			// Kategorisiere Score
			if contains(traditionalMethods, name) {
				traditionalScores[name] = score
			} else if contains(metadataMethods, name) {
				metadataScores[name] = score
			} else if contains(aiMethods, name) {
				aiScores[name] = score
			}

			// DEBUG: Log jeden Score
			fmt.Printf("RAW %s: %.3f\n", name, score)

			if name == "metadata" && score >= 0.95 {
				definitiveScore = 1.0
				reasoning = append(reasoning, "Definitive AI metadata found")
				fmt.Printf("DEFINITIVE: Metadata AI found (%.3f)\n", score)
				break
			}
			if name == "c2pa" && score >= 0.95 {
				definitiveScore = 1.0
				reasoning = append(reasoning, "C2PA certificate confirms AI generation")
				fmt.Printf("DEFINITIVE: C2PA AI found (%.3f)\n", score)
				break
			}
		}
	}

	traditionalAvg := calculateCategoryAverage(traditionalScores)
	metadataAvg := calculateCategoryAverage(metadataScores)
	aiAvg := calculateCategoryAverage(aiScores)

	if definitiveScore >= 0 {
		return createDefinitiveResponse(traditionalAvg, aiAvg, metadataAvg, scores, reasoning)
	}

	calibratedScores := applyBalancedCalibration(scores)

	fmt.Printf("\n=== CALIBRATED SCORES ===\n")
	for name, score := range calibratedScores {
		fmt.Printf("CAL %s: %.3f (was %.3f)\n", name, score, scores[name])
	}

	patternBoost := calculateAdvancedBoost(calibratedScores)
	fmt.Printf("\nPATTERN BOOST: %.3f\n", patternBoost)

	var weightedSum float64
	var totalWeight float64

	for name, score := range calibratedScores {
		weight := weights[name]
		if weight == 0 {
			continue
		}

		adaptiveWeight := weight
		if score >= 0.8 {
			adaptiveWeight *= 1.2
		} else if score <= 0.2 {
			adaptiveWeight *= 1.3
		}

		contribution := score * adaptiveWeight
		weightedSum += contribution
		totalWeight += adaptiveWeight

		fmt.Printf("CONTRIB %s: score=%.3f * weight=%.3f = %.3f\n",
			name, score, adaptiveWeight, contribution)

		if score >= 0.7 {
			reasoning = append(reasoning, fmt.Sprintf("%s: Strong AI indicators (%.0f%% probability)", name, score*100))
		} else if score <= 0.3 {
			reasoning = append(reasoning, fmt.Sprintf("%s: Authenticity indicators (%.0f%% authentic)", name, (1-score)*100))
		} else {
			reasoning = append(reasoning, fmt.Sprintf("%s: Moderate signals (%.0f%% probability)", name, score*100))
		}
	}

	if totalWeight == 0 {
		return map[string]interface{}{
			"verdict":     "Analysis Failed",
			"probability": 0.0,
			"confidence":  0.0,
			"summary":     "No usable analysis results obtained",
			"reasoning":   []string{"Technical error during analysis"},
			"scores":      scores,
		}
	}

	baseScore := weightedSum / totalWeight
	fmt.Printf("\nBASE SCORE: %.3f (weightedSum=%.3f / totalWeight=%.3f)\n",
		baseScore, weightedSum, totalWeight)

	baseScore *= patternBoost
	fmt.Printf("AFTER BOOST: %.3f\n", baseScore)

	analysisQuality := float64(len(scores)) / 10.0
	qualityBonus := 1.0
	if analysisQuality >= 0.8 {
		qualityBonus = 1.05
	} else if analysisQuality < 0.5 {
		qualityBonus = 0.95
	}

	finalScore := baseScore * qualityBonus
	fmt.Printf("FINAL SCORE: %.3f (quality=%.3f)\n", finalScore, qualityBonus)

	// Clamp auf 0-1 Bereich
	if finalScore > 1.0 {
		finalScore = 1.0
	} else if finalScore < 0.0 {
		finalScore = 0.0
	}

	verdict, confidence := determineBalancedVerdict(finalScore, calibratedScores)

	fmt.Printf("VERDICT: %s (%.1f%%)\n", verdict, finalScore*100)
	fmt.Printf("==================\n\n")

	return map[string]interface{}{
		"verdict":          verdict,
		"probability":      finalScore * 100,
		"confidence":       confidence,
		"summary":          fmt.Sprintf("%s - %.0f AI points with %.0f%% confidence", verdict, finalScore*100, confidence*100),
		"reasoning":        reasoning,
		"scores":           scores,
		"analysis_quality": calculateAnalysisQuality(len(results.Results), len(scores)),
		"quality_factor":   analysisQuality,

		"detailed_breakdown": map[string]interface{}{
			"traditional_computer_vision": map[string]interface{}{
				"average_score": traditionalAvg,
				"verdict":       getCategoryVerdict(traditionalAvg),
				"methods":       traditionalScores,
				"explanation":   generateTraditionalExplanation(traditionalScores),
			},
			"ai_deep_learning": map[string]interface{}{
				"average_score": aiAvg,
				"verdict":       getCategoryVerdict(aiAvg),
				"methods":       aiScores,
				"explanation":   generateAIExplanation(aiScores),
			},
			"metadata_forensics": map[string]interface{}{
				"average_score": metadataAvg,
				"verdict":       getCategoryVerdict(metadataAvg),
				"methods":       metadataScores,
				"explanation":   generateMetadataExplanation(metadataScores),
			},
			"method_agreement": analyzeMethodAgreement(traditionalAvg, aiAvg, metadataAvg),
		},
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func calculateCategoryAverage(scores map[string]float64) float64 {
	if len(scores) == 0 {
		return -1
	}

	var sum float64
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}

func createDefinitiveResponse(traditionalAvg, aiAvg, metadataAvg float64, scores map[string]float64, reasoning []string) map[string]interface{} {
	traditionalScores := make(map[string]float64)
	aiScores := make(map[string]float64)
	metadataScores := make(map[string]float64)

	// Re-kategorisiere für definitive Response
	traditionalMethods := []string{"artifacts", "compression", "pixel-analysis", "color-balance", "lighting-analysis", "advanced-artifacts", "object-coherence"}
	metadataMethods := []string{"metadata", "metadata-quick", "exif", "c2pa"}
	aiMethods := []string{"ai-model"}

	for name, score := range scores {
		if contains(traditionalMethods, name) {
			traditionalScores[name] = score
		} else if contains(metadataMethods, name) {
			metadataScores[name] = score
		} else if contains(aiMethods, name) {
			aiScores[name] = score
		}
	}

	return map[string]interface{}{
		"verdict":          "AI Generated (Confirmed)",
		"probability":      100.0,
		"confidence":       0.99,
		"summary":          "AI Generated (Confirmed) - 100.0% AI probability with 99% confidence",
		"reasoning":        reasoning,
		"scores":           scores,
		"analysis_quality": calculateAnalysisQuality(len(scores), len(scores)),
		"detailed_breakdown": map[string]interface{}{
			"traditional_computer_vision": map[string]interface{}{
				"average_score": traditionalAvg,
				"verdict":       getCategoryVerdict(traditionalAvg),
				"methods":       traditionalScores,
				"explanation":   generateTraditionalExplanation(traditionalScores),
			},
			"ai_deep_learning": map[string]interface{}{
				"average_score": aiAvg,
				"verdict":       getCategoryVerdict(aiAvg),
				"methods":       aiScores,
				"explanation":   generateAIExplanation(aiScores),
			},
			"metadata_forensics": map[string]interface{}{
				"average_score": metadataAvg,
				"verdict":       getCategoryVerdict(metadataAvg),
				"methods":       metadataScores,
				"explanation":   generateMetadataExplanation(metadataScores),
			},
			"method_agreement": analyzeMethodAgreement(traditionalAvg, aiAvg, metadataAvg),
		},
	}
}

func getCategoryVerdict(avgScore float64) string {
	if avgScore < 0 {
		return "No Data"
	} else if avgScore >= 0.7 {
		return "Strong AI Indicators"
	} else if avgScore >= 0.5 {
		return "Moderate AI Indicators"
	} else if avgScore >= 0.3 {
		return "Weak AI Indicators"
	} else {
		return "Authenticity Indicators"
	}
}

func calculateAnalysisQuality(totalAnalyses, successfulAnalyses int) string {
	if totalAnalyses == 0 {
		return "no_data"
	}

	successRate := float64(successfulAnalyses) / float64(totalAnalyses)

	if successRate >= 0.9 {
		return "excellent"
	} else if successRate >= 0.7 {
		return "good"
	} else if successRate >= 0.5 {
		return "fair"
	} else {
		return "poor"
	}
}
