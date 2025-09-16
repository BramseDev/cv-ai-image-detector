package pythonrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ScriptResult unverändert
type ScriptResult struct {
	Name string      `json:"name"`
	Data interface{} `json:"data,omitempty"`
	Err  string      `json:"err,omitempty"`
}

// RunMetadata ruft jetzt direkt exiftool -j auf.
func RunMetadata(ctx context.Context, imgPath string) (interface{}, error) {
	cmd := exec.CommandContext(ctx, "exiftool", "-j", imgPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exiftool failed: %v\n%s", err, out)
	}
	var meta []map[string]interface{}
	if err := json.Unmarshal(out, &meta); err != nil {
		return nil, err
	}
	// exiftool -j liefert eine Liste mit einem Objekt pro Datei
	if len(meta) > 0 {
		return meta[0], nil
	}
	return map[string]interface{}{}, nil
}

// RunCompression ruft analyze_compression.py auf und parst das JSON-Resultat.
func RunCompression(ctx context.Context, imgPath string) (interface{}, error) {
	script := filepath.Join("pythonScripts", "analyze_compression.py")

	cmd := exec.CommandContext(ctx, "python3", script, imgPath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start compression script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("compression analysis failed: %v\nStderr:\n%s", err, stderrBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, stdoutBytes)
	}
	return result, nil

}

func RunArtifacts(ctx context.Context, imagePath string) (interface{}, error) {
	scriptPath := filepath.Join("pythonScripts", "detect-artifacts.py")

	// Hier liegt der Fehler - wir müssen python3 als Befehl verwenden
	// und das Skript als Argument übergeben
	cmd := exec.CommandContext(ctx, "python3", scriptPath, imagePath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start artifacts detection script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("artifacts detection failed: %v\nStderr:\n%s", err, stderrBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, stdoutBytes)
	}
	return result, nil
}

// Füge diese Funktion zu deinen bestehenden runner.go-Funktionen hinzu

// RunColorBalance führt das Farbbalance-Analyse-Skript aus
func RunColorBalance(ctx context.Context, imgPath string) (interface{}, error) {
	script := filepath.Join("pythonScripts", "analyze_color_balance.py")

	cmd := exec.CommandContext(ctx, "python3", script, imgPath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start color balance script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("color balance analysis failed: %v\nStderr:\n%s", err, stderrBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, stdoutBytes)
	}
	return result, nil
}

func RunColorHistogram(ctx context.Context, imgPath string) (interface{}, error) {
	script := filepath.Join("pythonScripts", "analyze_color_histogram.py")

	cmd := exec.CommandContext(ctx, "python3", script, imgPath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start color histogram script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("color histogram analysis failed: %v\nStderr:\n%s", err, stderrBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, stdoutBytes)
	}
	return result, nil
}

func RunPixelAnalysis(ctx context.Context, imgPath string) (interface{}, error) {
	script := filepath.Join("pythonScripts", "analyze_pixel.py")

	cmd := exec.CommandContext(ctx, "python3", script, imgPath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start pixel analysis script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("pixel analysis failed: %v\nStderr:\n%s", err, stderrBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, stdoutBytes)
	}
	return result, nil
}

func RunAdvancedArtifacts(ctx context.Context, imagePath string) (interface{}, error) {
	scriptPath := filepath.Join("pythonScripts", "advanced-artifacts.py")

	cmd := exec.CommandContext(ctx, "python3", scriptPath, imagePath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start advanced artifacts script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("advanced artifacts failed: %v\nStderr:\n%s", err, stderrBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, stdoutBytes)
	}
	return result, nil
}
func RunObjectCoherence(ctx context.Context, imagePath string) (interface{}, error) {
	scriptPath := filepath.Join("pythonScripts", "analyze_coherence.py")

	cmd := exec.CommandContext(ctx, "python3", scriptPath, imagePath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start object coherence script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("object coherence analysis failed: %v\nStderr:\n%s", err, stderrBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, stdoutBytes)
	}
	return result, nil
}

func RunLightingAnalysis(ctx context.Context, imagePath string) (interface{}, error) {
	scriptPath := filepath.Join("pythonScripts", "analyze_lighting.py")

	cmd := exec.CommandContext(ctx, "python3", scriptPath, imagePath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start lighting analysis script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("lighting analysis failed: %v\nStderr:\n%s", err, stderrBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, stdoutBytes)
	}
	return result, nil
}

// func RunAIModelPrediction(ctx context.Context, imgPath string) (interface{}, error) {
// 	scriptPath := filepath.Join("ai-analyse", "modeltest.py")
// 	modelPath := filepath.Join("ai-analyse", "checkpoint.pt")

// 	// Prüfe ob Virtual Environment Python existiert
// 	pythonCmd := getPythonCommand()

// 	// Debug-Logs
// 	log.Printf("DEBUG: AI Model - Python: %s, Script: %s, Model: %s, Image: %s",
// 		pythonCmd, scriptPath, modelPath, imgPath)

// 	// Prüfe ob Dateien existieren
// 	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
// 		return nil, fmt.Errorf("AI model script not found: %s", scriptPath)
// 	}
// 	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
// 		return nil, fmt.Errorf("AI model checkpoint not found: %s", modelPath)
// 	}

// 	cmd := exec.CommandContext(ctx, pythonCmd, scriptPath,
// 		"--model", modelPath,
// 		"--image", imgPath)

// 	stdoutPipe, _ := cmd.StdoutPipe()
// 	stderrPipe, _ := cmd.StderrPipe()

// 	if err := cmd.Start(); err != nil {
// 		return nil, fmt.Errorf("could not start AI model script: %v", err)
// 	}

// 	stdoutBytes, _ := io.ReadAll(stdoutPipe)
// 	stderrBytes, _ := io.ReadAll(stderrPipe)

// 	if err := cmd.Wait(); err != nil {
// 		return nil, fmt.Errorf("AI model prediction failed: %v\nStderr:\n%s", err, string(stderrBytes))
// 	}

// 	// Parse JSON direkt (modeltest.py gibt bereits JSON zurück)
// 	var result map[string]interface{}
// 	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
// 		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, string(stdoutBytes))
// 	}

//		log.Printf("DEBUG: AI Model Result: %+v", result)
//		return result, nil
//	}

func parseClassifyV6Output(output, imgPath string) (interface{}, error) {
	log.Printf("DEBUG: Full classify-v6 output:\n%s", output)

	lines := strings.Split(output, "\n")
	var predLine string

	// Find prediction line - flexiblere Suche
	imageName := filepath.Base(imgPath)
	log.Printf("DEBUG: Looking for image name: %s", imageName)

	for _, line := range lines {
		log.Printf("DEBUG: Checking line: %s", line)
		if strings.Contains(line, "img:") {
			predLine = line
			log.Printf("DEBUG: Found prediction line: %s", predLine)
			break
		}
	}

	if predLine == "" {
		log.Printf("DEBUG: No prediction line found in output")
		return nil, fmt.Errorf("no prediction found in output")
	}

	// Parse: "img: KI-Bilder-erstellen.jpg pred: FAKE prob: 1.000 conf: 1.000"
	parts := strings.Fields(predLine)
	log.Printf("DEBUG: Parsed fields: %v", parts)

	if len(parts) < 8 {
		return nil, fmt.Errorf("invalid prediction format: %s (parts: %d)", predLine, len(parts))
	}

	var prediction string
	var probability, confidence float64
	var err error

	// Robusteres Parsing
	for i, part := range parts {
		switch part {
		case "pred:":
			if i+1 < len(parts) {
				prediction = parts[i+1]
				log.Printf("DEBUG: Found prediction: %s", prediction)
			}
		case "prob:":
			if i+1 < len(parts) {
				if probability, err = strconv.ParseFloat(parts[i+1], 64); err != nil {
					log.Printf("DEBUG: Error parsing probability: %v", err)
					return nil, fmt.Errorf("invalid probability value: %v", err)
				}
				log.Printf("DEBUG: Found probability: %f", probability)
			}
		case "conf:":
			if i+1 < len(parts) {
				if confidence, err = strconv.ParseFloat(parts[i+1], 64); err != nil {
					log.Printf("DEBUG: Error parsing confidence: %v", err)
					return nil, fmt.Errorf("invalid confidence value: %v", err)
				}
				log.Printf("DEBUG: Found confidence: %f", confidence)
			}
		}
	}

	if prediction == "" {
		return nil, fmt.Errorf("could not parse prediction from: %s", predLine)
	}

	// Convert to expected format
	isFake := prediction == "FAKE"
	authenticityScore := probability
	if isFake {
		// Für FAKE: higher prob means more AI-like, so lower authenticity
		authenticityScore = 1.0 - probability
	}

	result := map[string]interface{}{
		"prediction":  strings.ToLower(prediction),
		"probability": probability,
		"confidence":  confidence,
		"model_type":  "ensemble_efficientnetv2",
		"ai_model_analysis": map[string]interface{}{
			"predicted_class":    strings.ToLower(prediction),
			"confidence_score":   confidence,
			"is_ai_generated":    isFake,
			"authenticity_score": authenticityScore,
		},
	}

	log.Printf("DEBUG: Final parsed result: %+v", result)
	return result, nil
}

// // Helper function to get the correct Python command
// func getPythonCommand() string {
// 	// Prüfe Windows Virtual Environment zuerst
// 	venvPythonWin := filepath.Join("venv", "Scripts", "python.exe")
// 	if _, err := os.Stat(venvPythonWin); err == nil {
// 		return venvPythonWin
// 	}

// 	// Prüfe Unix Virtual Environment (venv/bin/python3)
// 	venvPython := filepath.Join("venv", "bin", "python3")
// 	if _, err := os.Stat(venvPython); err == nil {
// 		return venvPython
// 	}

// 	// Prüfe alternative Unix Virtual Environment Pfade
// 	venvPython2 := filepath.Join("venv", "bin", "python")
// 	if _, err := os.Stat(venvPython2); err == nil {
// 		return venvPython2
// 	}

//		// Fallback zu System-Python
//		return "python3"
//	}
func getPythonCommand() string {
	// Nur die funktionierende Virtual Environment verwenden
	venvPython := filepath.Join("venv", "bin", "python3")

	// Test if the venv python works with required modules
	if _, err := os.Stat(venvPython); err == nil {
		testCmd := exec.Command(venvPython, "-c", "import timm, torch; print('OK')")
		if err := testCmd.Run(); err == nil {
			log.Printf("DEBUG: Using working venv Python: %s", venvPython)
			return venvPython
		} else {
			log.Printf("WARNING: venv Python found but modules missing: %v", err)
		}
	}

	log.Printf("ERROR: No working Python environment found!")
	return "python3" // Fallback
}

func findWorkingPython() string {
	// Try venv Python first
	venvPython := filepath.Join("venv", "bin", "python3")
	if _, err := os.Stat(venvPython); err == nil {
		testCmd := exec.Command(venvPython, "-c", "import timm, torch; print('OK')")
		if err := testCmd.Run(); err == nil {
			log.Printf("DEBUG: Using working venv Python: %s", venvPython)
			return venvPython
		} else {
			log.Printf("WARNING: venv Python found but modules missing: %v", err)
		}
	}

	log.Printf("ERROR: No working Python environment found!")
	return "python3" // Fallback
}

func RunAIModelPrediction(ctx context.Context, imgPath string) (interface{}, error) {
	scriptPath := filepath.Join("ai-analyse", "new_analysis", "classify-v6.py")
	modelsDir := filepath.Join("ai-analyse", "new_analysis", "ensemble1")

	pythonCmd := findWorkingPython()

	cmd := exec.CommandContext(ctx, pythonCmd, scriptPath,
		"--models", modelsDir,
		imgPath)

	cmd.Dir = "."

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Printf("AI model script stderr: %s", string(exitError.Stderr))
		}
		return nil, fmt.Errorf("AI model prediction failed: %v", err)
	}

	log.Printf("AI model raw output: %s", string(output))

	return parseClassifyV6Output(string(output), imgPath)
}
