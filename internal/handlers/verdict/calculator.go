package verdict

import (
	"fmt"
	"math"
	"strings"

	"github.com/BramseDev/imageAnalyzer/pkg/analyzer/pipeline"
)

func CalculateOverallVerdict(results *pipeline.PipelineResult) map[string]interface{} {
	scores := make(map[string]float64)
	reasoning := []string{}

	tempPipeline := pipeline.NewAnalysisPipeline()

	// Methoden-Kategorisierung - METADATA UND TRADITIONAL ZUSAMMENGEFASST
	computerVisionMethods := []string{
		"artifacts",
		"lighting-analysis",
		"advanced-artifacts",
		"pixel-analysis",
		"color-balance",
		"object-coherence",
		"compression",
		"metadata",
		"c2pa",
		"exif",
		"metadata-quick",
	}

	aiMethods := []string{
		"ai-model",
	}

	// Separate Score-Sammlung - NUR NOCH 2 KATEGORIEN
	computerVisionScores := make(map[string]float64)
	aiScores := make(map[string]float64)

	// DEBUG: Log alle rohen Scores
	fmt.Printf("\n=== DEBUG SCORES ===\n")

	weights := map[string]float64{
		"ai-model": 6.0,

		"compression":        4.0,
		"lighting-analysis":  3.5,
		"artifacts":          3.0,
		"advanced-artifacts": 3.0,
		"color-balance":      3.0,

		"metadata":       2.5,
		"pixel-analysis": 2.5,
		"c2pa":           2.0,

		"object-coherence": 0.5,
		"exif":             1.0,
		"metadata-quick":   0.8,
	}

	var definitiveScore float64 = -1

	for name, result := range results.Results {
		var score float64 = -1

		// Type assertion zu map[string]interface{}
		resultData, ok := result.(map[string]interface{})
		if !ok {
			fmt.Printf("WARNING: %s has invalid data type: %T\n", name, result)
			continue
		}

		// Verwende spezialisierte Score-Funktionen
		switch name {
		case "artifacts":
			score = calculateArtifactsScore(resultData)
		case "advanced-artifacts":
			score = calculateAdvancedArtifactsScore(resultData)
		case "ai-model":
			score = calculateAIModelScore(resultData)
		case "compression":
			score = calculateCompressionScore(resultData)
		case "pixel-analysis":
			score = calculatePixelAnalysisScore(resultData)
		case "lighting-analysis":
			score = calculateLightingAnalysisScore(resultData)
		case "color-balance":
			score = calculateColorBalanceScore(resultData)
		case "object-coherence":
			score = calculateObjectCoherenceScore(resultData)
		case "exif":
			score = calculateEXIFScore(resultData)
		case "metadata":
			score = calculateMetadataScore(resultData)
		case "metadata-quick":
			score = calculateMetadataQuickScore(resultData)
		case "c2pa":
			score = calculateC2PAScore(resultData)
		default:
			// Fallback zur Pipeline
			score = tempPipeline.ExtractConfidenceScore(result)
		}

		if score >= 0 {
			scores[name] = score

			// Kategorisiere Score - NUR NOCH 2 KATEGORIEN
			if contains(computerVisionMethods, name) {
				computerVisionScores[name] = score
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

	// Berechne separate Scores - NUR NOCH 2 KATEGORIEN
	computerVisionScore := calculateComputerVisionScore(computerVisionScores)
	aiAnalysisScore := calculateAIAnalysisScore(aiScores)

	// DEBUG: Log separate scores
	fmt.Printf("\n=== SEPARATE SCORES ===\n")
	fmt.Printf("Computer Vision (incl. Metadata): %.3f\n", computerVisionScore)
	fmt.Printf("AI Deep Learning: %.3f\n", aiAnalysisScore)

	var finalScore float64
	var verdict string
	var confidence float64

	if definitiveScore >= 0 {
		finalScore = definitiveScore
		verdict = "AI Generated"
		confidence = 0.95
	} else {
		// Kalibrierte Scores
		calibratedScores := applyBalancedCalibration(scores)

		fmt.Printf("\n=== CALIBRATED SCORES ===\n")
		for name, score := range calibratedScores {
			fmt.Printf("CAL %s: %.3f (was %.3f)\n", name, score, scores[name])
		}

		// Pattern-Boost
		patternBoost := calculateAdvancedBoost(calibratedScores)
		fmt.Printf("\nPATTERN BOOST: %.3f\n", patternBoost)

		var weightedSum float64
		var totalWeight float64

		for name, score := range calibratedScores {
			weight := weights[name]
			if weight == 0 {
				continue
			}

			if name == "ai-model" {
				continue // Skip AI-Model für finalScore
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

		finalScore = baseScore * qualityBonus
		fmt.Printf("FINAL SCORE: %.3f (quality=%.3f)\n", finalScore, qualityBonus)

		// Clamp auf 0-1 Bereich
		if finalScore > 1.0 {
			finalScore = 1.0
		} else if finalScore < 0.0 {
			finalScore = 0.0
		}

		verdict, confidence = determineBalancedVerdict(finalScore, calibratedScores)
	}

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

		// NEUE SEPARATE BEWERTUNGEN - NUR NOCH 2 KATEGORIEN
		"separate_analysis": map[string]interface{}{
			"computer_vision": map[string]interface{}{
				"score":       computerVisionScore,
				"percentage":  computerVisionScore * 100,
				"methods":     computerVisionScores,
				"verdict":     getCategoryVerdict(computerVisionScore),
				"explanation": generateComputerVisionExplanation(computerVisionScores),
			},
			"ai_analysis": map[string]interface{}{
				"score":       aiAnalysisScore,
				"percentage":  aiAnalysisScore * 100,
				"methods":     aiScores,
				"verdict":     getCategoryVerdict(aiAnalysisScore),
				"explanation": generateAIExplanation(aiScores),
			},
			"comparison": map[string]interface{}{
				"cv_vs_ai_difference": calculateDifference(computerVisionScore, aiAnalysisScore),
				"agreement_level":     calculateAgreementLevel(computerVisionScore, aiAnalysisScore),
				"dominant_method":     getDominantMethodSimple(computerVisionScore, aiAnalysisScore),
			},
		},

		"detailed_breakdown": map[string]interface{}{
			"weighted_scores": calculateWeightedBreakdown(scores, weights),
			"method_groups": map[string]interface{}{
				"computer_vision":  computerVisionScores, // Enthält jetzt auch Metadata
				"ai_deep_learning": aiScores,
			},
			"strength_indicators": analyzeStrengthIndicators(scores),
			"consistency_check":   checkConsistency(scores),
		},
	}
}

// Helper-Funktionen für separate Score-Berechnung
func calculateComputerVisionScore(computerVisionScores map[string]float64) float64 {
	if len(computerVisionScores) == 0 {
		return -1
	}

	var sum float64
	for _, score := range computerVisionScores {
		sum += score
	}
	return sum / float64(len(computerVisionScores))
}

func calculateAIAnalysisScore(aiScores map[string]float64) float64 {
	if len(aiScores) == 0 {
		return -1
	}

	var sum float64
	for _, score := range aiScores {
		sum += score
	}
	return sum / float64(len(aiScores))
}

func getCategoryVerdict(score float64) string {
	if score < 0 {
		return "No Data"
	} else if score >= 0.7 {
		return "AI Generated"
	} else if score >= 0.3 {
		return "Likely Human"
	} else {
		return "Likely Human"
	}
}

func calculateDifference(score1, score2 float64) float64 {
	if score1 < 0 || score2 < 0 {
		return -1
	}
	return math.Abs(score1 - score2)
}

func calculateAgreementLevel(cvScore, aiScore float64) string {
	if cvScore < 0 || aiScore < 0 {
		return "insufficient_data"
	}

	diff := math.Abs(cvScore - aiScore)
	if diff <= 0.1 {
		return "strong_agreement"
	} else if diff <= 0.3 {
		return "moderate_agreement"
	} else if diff <= 0.5 {
		return "weak_agreement"
	} else {
		return "strong_disagreement"
	}
}

func getDominantMethodSimple(cvScore, aiScore float64) string {
	if cvScore < 0 && aiScore < 0 {
		return "no_data"
	} else if cvScore < 0 {
		return "ai_analysis"
	} else if aiScore < 0 {
		return "computer_vision"
	} else if cvScore > aiScore {
		return "computer_vision"
	} else if aiScore > cvScore {
		return "ai_analysis"
	} else {
		return "equal"
	}
}

func generateComputerVisionExplanation(scores map[string]float64) string {
	if len(scores) == 0 {
		return "No computer vision or metadata analysis available"
	}

	var explanations []string
	for method, score := range scores {
		if score >= 0.7 {
			if method == "metadata" || method == "c2pa" || method == "exif" {
				explanations = append(explanations, fmt.Sprintf("%s found strong AI markers (%.1f%%)", method, score*100))
			} else {
				explanations = append(explanations, fmt.Sprintf("%s indicates strong AI artifacts (%.1f%%)", method, score*100))
			}
		} else if score >= 0.3 {
			explanations = append(explanations, fmt.Sprintf("%s shows mixed signals (%.1f%%)", method, score*100))
		} else {
			if method == "metadata" || method == "c2pa" || method == "exif" {
				explanations = append(explanations, fmt.Sprintf("%s found clean metadata (%.1f%%)", method, score*100))
			} else {
				explanations = append(explanations, fmt.Sprintf("%s suggests human origin (%.1f%%)", method, score*100))
			}
		}
	}

	return strings.Join(explanations, "; ")
}

// Bestehende Helper-Funktionen
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func calculateAnalysisQuality(totalMethods, successfulMethods int) float64 {
	if totalMethods == 0 {
		return 0.0
	}
	return float64(successfulMethods) / float64(totalMethods)
}

func calculateWeightedBreakdown(scores map[string]float64, weights map[string]float64) map[string]float64 {
	breakdown := make(map[string]float64)
	for name, score := range scores {
		if weight, exists := weights[name]; exists {
			breakdown[name] = score * weight
		}
	}
	return breakdown
}

func analyzeStrengthIndicators(scores map[string]float64) []string {
	indicators := []string{}

	for name, score := range scores {
		if score >= 0.9 {
			indicators = append(indicators, fmt.Sprintf("Very strong: %s", name))
		} else if score >= 0.7 {
			indicators = append(indicators, fmt.Sprintf("Strong: %s", name))
		}
	}

	return indicators
}

func checkConsistency(scores map[string]float64) map[string]interface{} {
	if len(scores) < 2 {
		return map[string]interface{}{
			"level":      "insufficient_data",
			"variance":   0.0,
			"assessment": "Need more methods for consistency check",
		}
	}

	// Berechne Varianz
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

	var level string
	var assessment string

	if variance <= 0.1 {
		level = "high"
		assessment = "Methods show strong agreement"
	} else if variance <= 0.3 {
		level = "moderate"
		assessment = "Methods show reasonable consistency"
	} else {
		level = "low"
		assessment = "Methods show significant disagreement"
	}

	return map[string]interface{}{
		"level":      level,
		"variance":   variance,
		"mean":       mean,
		"assessment": assessment,
	}
}

func generateAIExplanation(scores map[string]float64) string {
	if len(scores) == 0 {
		return "No AI deep learning analysis available"
	}

	for _, score := range scores {
		if score >= 0.9 {
			return fmt.Sprintf("Neural network strongly predicts AI-generated (%.1f%% confidence). Very high certainty from deep learning model.", score*100)
		} else if score >= 0.7 {
			return fmt.Sprintf("Neural network indicates likely AI-generated (%.1f%% confidence). Strong AI detection signals.", score*100)
		} else if score >= 0.3 {
			return fmt.Sprintf("Neural network shows mixed results (%.1f%% confidence). Uncertain classification from AI model.", score*100)
		} else {
			return fmt.Sprintf("Neural network suggests human origin (%.1f%% authentic). Low AI detection confidence.", (1-score)*100)
		}
	}

	return "AI deep learning analysis completed"
}
