package rustrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

type ScriptResult struct {
	Name string      `json:"name"`
	Data interface{} `json:"data,omitempty"`
	Err  string      `json:"err,omitempty"`
}

func RunC2PA(ctx context.Context, imgPath string) (interface{}, error) {
	binaryPath := filepath.Join("pkg", "analyzer", "c2pa-rust", "target", "release", "c2pa-rust")

	cmd := exec.CommandContext(ctx, binaryPath, imgPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("c2pa-rust failed: %v\n%s", err, out)
	}

	var result interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %w\nRaw output: %s", err, out)
	}

	return result, nil
}
