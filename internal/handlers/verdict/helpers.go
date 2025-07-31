package verdict

import (
	"fmt"
	"strings"
)

// Traditional Methods Erklärung (aus alter upload.go)
func generateTraditionalExplanation(scores map[string]float64) string {
	highScores := []string{}
	lowScores := []string{}

	methodNames := map[string]string{
		"artifacts":          "Visual Artifacts",
		"compression":        "Compression Analysis",
		"pixel-analysis":     "Pixel-Level Analysis",
		"color-balance":      "Color Distribution",
		"lighting-analysis":  "Lighting Physics",
		"advanced-artifacts": "Mathematical Patterns",
		"object-coherence":   "Object Consistency",
	}

	for method, score := range scores {
		displayName := methodNames[method]
		if displayName == "" {
			displayName = method
		}

		if score >= 0.6 {
			highScores = append(highScores, displayName)
		} else if score <= 0.3 {
			lowScores = append(lowScores, displayName)
		}
	}

	if len(highScores) > 0 {
		return fmt.Sprintf("Computer Vision detected AI patterns in: %s", strings.Join(highScores, ", "))
	} else if len(lowScores) > 0 {
		return fmt.Sprintf("Computer Vision found authentic patterns in: %s", strings.Join(lowScores, ", "))
	} else {
		return "Computer Vision analysis shows mixed results"
	}
}

// AI Model Erklärung (aus alter upload.go)
func generateAIExplanation(scores map[string]float64) string {
	if aiScore, exists := scores["ai-model"]; exists {
		if aiScore >= 0.8 {
			return fmt.Sprintf("Neural network strongly predicts AI-generated (%.1f%% confidence). Very high certainty from deep learning model.", aiScore*100)
		} else if aiScore >= 0.7 {
			return fmt.Sprintf("Neural network predicts AI-generated (%.1f%% confidence). High certainty from ResNet-18 model.", aiScore*100)
		} else if aiScore >= 0.5 {
			return fmt.Sprintf("Neural network leans towards AI-generated (%.1f%% confidence). Moderate certainty from trained model.", aiScore*100)
		} else if aiScore <= 0.2 {
			return fmt.Sprintf("Neural network strongly predicts authentic (%.1f%% confidence). Deep learning model very confident in real photography.", (1-aiScore)*100)
		} else if aiScore <= 0.3 {
			return fmt.Sprintf("Neural network predicts authentic (%.1f%% confidence). Model suggests real camera origin.", (1-aiScore)*100)
		} else {
			return fmt.Sprintf("Neural network uncertain (%.1f%% towards AI). Model shows mixed signals.", aiScore*100)
		}
	}
	return "No AI model analysis available"
}

// Metadata Erklärung (aus alter upload.go)
func generateMetadataExplanation(scores map[string]float64) string {
	findings := []string{}

	for method, score := range scores {
		switch method {
		case "c2pa":
			if score >= 0.9 {
				findings = append(findings, "C2PA certificate confirms AI generation")
			} else if score <= 0.1 {
				findings = append(findings, "No C2PA AI markers found")
			} else {
				findings = append(findings, "C2PA analysis inconclusive")
			}
		case "exif":
			if score <= 0.1 {
				findings = append(findings, "Rich EXIF data suggests camera origin")
			} else if score >= 0.5 {
				findings = append(findings, "Suspicious EXIF patterns detected")
			} else {
				findings = append(findings, "Standard EXIF data present")
			}
		case "metadata":
			if score <= 0.2 {
				findings = append(findings, "Complete technical metadata present")
			} else if score >= 0.8 {
				findings = append(findings, "Metadata indicates AI generation")
			} else {
				findings = append(findings, "Standard metadata analysis")
			}
		case "metadata-quick":
			// Meist weniger relevant für Erklärung
			continue
		}
	}

	if len(findings) == 0 {
		return "Standard metadata analysis completed"
	}
	return strings.Join(findings, "; ")
}

func analyzeMethodAgreement(traditional, ai, metadata float64) map[string]interface{} {
	agreement := map[string]interface{}{
		"overall_consensus":  "mixed",
		"conflicts":          []string{},
		"agreements":         []string{},
		"reliability_score":  0.5,
		"consensus_strength": "weak",
	}

	// Definiere Schwellenwerte
	highThreshold := 0.6
	lowThreshold := 0.3

	methods := map[string]float64{
		"Traditional Computer Vision": traditional,
		"AI Deep Learning":            ai,
		"Metadata Forensics":          metadata,
	}

	// Zähle High/Low Scores
	highCount := 0
	lowCount := 0
	highMethods := []string{}
	lowMethods := []string{}

	for name, score := range methods {
		if score >= 0 { // Only count methods that have data
			if score >= highThreshold {
				highCount++
				highMethods = append(highMethods, name)
			} else if score <= lowThreshold {
				lowCount++
				lowMethods = append(lowMethods, name)
			}
		}
	}

	// Analysiere Konsens
	totalMethods := 0
	for _, score := range methods {
		if score >= 0 {
			totalMethods++
		}
	}

	if highCount >= 2 && totalMethods >= 2 {
		agreement["overall_consensus"] = "ai_likely"
		agreement["reliability_score"] = 0.8 + (float64(highCount) * 0.05)
		agreement["consensus_strength"] = "strong"
		for _, method := range highMethods {
			agreement["agreements"] = append(agreement["agreements"].([]string),
				fmt.Sprintf(" %s detects AI indicators", method))
		}
	} else if lowCount >= 2 && totalMethods >= 2 {
		agreement["overall_consensus"] = "authentic_likely"
		agreement["reliability_score"] = 0.8 + (float64(lowCount) * 0.05)
		agreement["consensus_strength"] = "strong"
		for _, method := range lowMethods {
			agreement["agreements"] = append(agreement["agreements"].([]string),
				fmt.Sprintf(" %s finds authenticity signs", method))
		}
	} else {
		agreement["overall_consensus"] = "conflicting"
		agreement["reliability_score"] = 0.3 + (float64(totalMethods) * 0.1)
		agreement["consensus_strength"] = "weak"

		// Identifiziere spezifische Konflikte
		if traditional >= highThreshold && ai >= 0 && ai <= lowThreshold {
			agreement["conflicts"] = append(agreement["conflicts"].([]string),
				"Computer Vision detects AI, but Deep Learning suggests authentic")
		}
		if ai >= highThreshold && traditional >= 0 && traditional <= lowThreshold {
			agreement["conflicts"] = append(agreement["conflicts"].([]string),
				"Deep Learning detects AI, but Computer Vision suggests authentic")
		}
		if metadata >= 0 && metadata <= lowThreshold && (traditional >= highThreshold || ai >= highThreshold) {
			agreement["conflicts"] = append(agreement["conflicts"].([]string),
				"Clean metadata conflicts with analysis results")
		}
		if metadata >= highThreshold && (traditional <= lowThreshold || ai <= lowThreshold) {
			agreement["conflicts"] = append(agreement["conflicts"].([]string),
				"Suspicious metadata conflicts with analysis results")
		}

		// Füge positive Übereinstimmungen hinzu auch bei Konflikten
		if len(highMethods) > 0 {
			for _, method := range highMethods {
				agreement["agreements"] = append(agreement["agreements"].([]string),
					fmt.Sprintf("%s shows AI indicators", method))
			}
		}
		if len(lowMethods) > 0 {
			for _, method := range lowMethods {
				agreement["agreements"] = append(agreement["agreements"].([]string),
					fmt.Sprintf("%s shows authenticity signs", method))
			}
		}
	}

	return agreement
}
