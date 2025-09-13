package verdict

import (
	"fmt"
	"strings"
)

// Traditional Methods Erklärung (aus alter upload.go)
func generateTraditionalExplanation(scores map[string]float64) string {
	highScores := []string{}
	lowScores := []string{}

	for method, score := range scores {
		if score >= 0.7 {
			highScores = append(highScores, method)
		} else if score <= 0.3 {
			lowScores = append(lowScores, method)
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

// ENTFERNE DIESE FUNKTION - sie ist bereits in calculator.go definiert
// func generateAIExplanation(scores map[string]float64) string { ... }

// Metadata Erklärung (aus alter upload.go)
func generateMetadataExplanation(scores map[string]float64) string {
	findings := []string{}

	for method, score := range scores {
		switch method {
		case "c2pa":
			if score >= 0.7 {
				findings = append(findings, "C2PA verification indicates AI generation")
			} else {
				findings = append(findings, "No C2PA AI markers found")
			}
		case "exif":
			if score >= 0.5 {
				findings = append(findings, "EXIF metadata shows processing indicators")
			} else {
				findings = append(findings, "EXIF metadata appears natural")
			}
		case "metadata":
			if score >= 0.7 {
				findings = append(findings, "General metadata analysis suggests AI origin")
			} else {
				findings = append(findings, "Metadata patterns appear authentic")
			}
		}
	}

	if len(findings) == 0 {
		return "No metadata analysis available"
	}

	return strings.Join(findings, "; ")
}

func analyzeMethodAgreement(traditional, ai, metadata float64) map[string]interface{} {
	agreement := "unknown"
	confidence := 0.5

	// Wenn alle 3 Kategorien ähnliche Scores haben
	maxDiff := 0.0
	scores := []float64{traditional, ai, metadata}
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i] >= 0 && scores[j] >= 0 {
				diff := abs(scores[i] - scores[j])
				if diff > maxDiff {
					maxDiff = diff
				}
			}
		}
	}

	if maxDiff <= 0.2 {
		agreement = "strong"
		confidence = 0.9
	} else if maxDiff <= 0.4 {
		agreement = "moderate"
		confidence = 0.7
	} else {
		agreement = "weak"
		confidence = 0.5
	}

	return map[string]interface{}{
		"agreement_level": agreement,
		"confidence":      confidence,
		"max_difference":  maxDiff,
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
