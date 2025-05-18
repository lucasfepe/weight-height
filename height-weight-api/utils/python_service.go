package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/lucasfepe/height-weight-api/models"
)

// ML service URL
var mlServiceURL = "http://localhost:5000/predict"

// SetMLServiceURL allows changing the ML service URL at runtime
func SetMLServiceURL(url string) {
	mlServiceURL = url
}

// CallMLService sends the image to the ML service and returns the estimation results
func CallMLService(imageBytes []byte) (*models.MLServiceResponse, error) {
	// Create a new HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Prepare the request
	req, err := http.NewRequest("POST", mlServiceURL, bytes.NewBuffer(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "multipart/form-data")

	// Create multipart form writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file
	part, err := writer.CreateFormFile("image", "image.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Write image bytes to form
	if _, err := io.Copy(part, bytes.NewReader(imageBytes)); err != nil {
		return nil, fmt.Errorf("failed to copy image to form: %w", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Update the request with the form body
	req, err = http.NewRequest("POST", mlServiceURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request with form: %w", err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to ML service: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from ML service: %w", err)
	}

	// Check for non-200 response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result models.MLServiceResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ML service response: %w, response: %s", err, string(respBody))
	}

	return &result, nil
}

// For backward compatibility, if needed
func CallPythonML(imagePath, scriptPath string) (*models.MLServiceResponse, error) {
	// Read the image file
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	// Call the ML service
	return CallMLService(imageBytes)
}
