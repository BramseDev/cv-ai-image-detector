package explanations

import (
	"fmt"

	"github.com/BramseDev/imageAnalyzer/internal/handlers/utils"
)

func GenerateEXIFExplanation(data map[string]interface{}) string {
	hasCameraInfo, _ := utils.GetFloatValue(data, "has_camera_info")

	if hasCameraInfo > 0 {
		cameraMake, _ := utils.GetStringValue(data, "camera_make")
		cameraModel, _ := utils.GetStringValue(data, "camera_model")

		if cameraMake != "" && cameraModel != "" {
			return fmt.Sprintf("Complete camera metadata found: %s %s. Strong authenticity indicator.", cameraMake, cameraModel)
		}
		return "Camera metadata present. Indicates likely authentic photograph."
	}

	return "No camera-specific EXIF data found. Common in AI-generated images or heavily processed photos."
}
