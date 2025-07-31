package utils

import (
	"crypto/rand"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize       = 50 * 1024 * 1024 // 50MB
	MaxFilenameLength = 255
)

func ValidateFile(header *multipart.FileHeader) error {
	if header.Size > MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d)", header.Size, MaxFileSize)
	}

	if len(header.Filename) > MaxFilenameLength {
		return fmt.Errorf("filename too long: %d characters (max: %d)", len(header.Filename), MaxFilenameLength)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp"}

	validExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}

	if !validExt {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	return nil
}

func ValidateFileContent(tempFile string) error {
	file, err := os.Open(tempFile)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("unable to analyze file content")
	}

	contentType := http.DetectContentType(buffer)
	validTypes := []string{
		"image/jpeg", "image/png", "image/gif",
		"image/webp", "image/bmp", "image/tiff",
	}

	validContentType := false
	for _, validType := range validTypes {
		if contentType == validType {
			validContentType = true
			break
		}
	}

	if !validContentType {
		return fmt.Errorf("file content does not match expected image format. Detected type: %s", contentType)
	}

	file.Seek(0, 0)

	_, _, err = image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("invalid image file: %w", err)
	}

	return nil
}

func CreateSecureTempFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Generate secure random filename
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random filename: %w", err)
	}

	tempFile := fmt.Sprintf("/tmp/analyzer_%x%s", randomBytes, filepath.Ext(header.Filename))

	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return tempFile, nil
}
