package docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// BuildImage builds a Docker image from Dockerfile
func BuildImage(imageName, tag, contextPath string) error {
	// Check if Dockerfile exists
	dockerfilePath := filepath.Join(contextPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err != nil {
		return fmt.Errorf("Dockerfile not found in %s", contextPath)
	}

	imageTag := fmt.Sprintf("%s:%s", imageName, tag)

	// Build image
	cmd := exec.Command("docker", "build", "-t", imageTag, contextPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker build failed: %w", err)
	}

	return nil
}

// SaveImage saves a Docker image to a tar file
func SaveImage(imageName, tag, outputPath string) error {
	imageTag := fmt.Sprintf("%s:%s", imageName, tag)

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save image to tar
	cmd := exec.Command("docker", "save", "-o", outputPath, imageTag)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

// LoadImage loads a Docker image from a tar file on remote server via SSH
func LoadImage(sshClient interface{}, tarPath string) error {
	// This will be called via SSH
	// For now, just generate the command
	// Actual execution happens in deploy package
	return nil
}

// GetImageSize returns the size of a tar file in MB
func GetImageSize(tarPath string) (float64, error) {
	info, err := os.Stat(tarPath)
	if err != nil {
		return 0, err
	}

	sizeMB := float64(info.Size()) / 1024.0 / 1024.0
	return sizeMB, nil
}

// StreamBuild builds image and streams output
func StreamBuild(imageName, tag, contextPath string, stdout, stderr io.Writer) error {
	dockerfilePath := filepath.Join(contextPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err != nil {
		return fmt.Errorf("Dockerfile not found in %s", contextPath)
	}

	imageTag := fmt.Sprintf("%s:%s", imageName, tag)

	cmd := exec.Command("docker", "build", "-t", imageTag, contextPath)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

