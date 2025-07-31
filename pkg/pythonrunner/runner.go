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

func RunAIModelPrediction(ctx context.Context, imgPath string) (interface{}, error) {
	scriptPath := filepath.Join("ai-analyse", "modeltest.py")
	modelPath := filepath.Join("ai-analyse", "checkpoint.pt")

	// Prüfe ob Virtual Environment Python existiert
	pythonCmd := getPythonCommand()

	// Debug-Logs
	log.Printf("DEBUG: AI Model - Python: %s, Script: %s, Model: %s, Image: %s",
		pythonCmd, scriptPath, modelPath, imgPath)

	// Prüfe ob Dateien existieren
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("AI model script not found: %s", scriptPath)
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("AI model checkpoint not found: %s", modelPath)
	}

	cmd := exec.CommandContext(ctx, pythonCmd, scriptPath,
		"--model", modelPath,
		"--image", imgPath)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start AI model script: %v", err)
	}

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("AI model prediction failed: %v\nStderr:\n%s", err, string(stderrBytes))
	}

	// Parse JSON direkt (modeltest.py gibt bereits JSON zurück)
	var result map[string]interface{}
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v\nStdout:\n%s", err, string(stdoutBytes))
	}

	log.Printf("DEBUG: AI Model Result: %+v", result)
	return result, nil
}

// Helper function to get the correct Python command
func getPythonCommand() string {
	// Prüfe Virtual Environment zuerst
	venvPython := filepath.Join("venv", "bin", "python3")
	if _, err := os.Stat(venvPython); err == nil {
		return venvPython
	}

	// Prüfe alternative Virtual Environment Pfade
	venvPython2 := filepath.Join("venv", "bin", "python")
	if _, err := os.Stat(venvPython2); err == nil {
		return venvPython2
	}

	// Fallback zu System-Python
	return "python3"
}
