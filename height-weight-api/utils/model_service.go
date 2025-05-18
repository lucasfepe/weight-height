package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/lucasfepe/height-weight-api/config"
	"github.com/lucasfepe/height-weight-api/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ModelResponse represents the response from the TensorFlow model service
type ModelResponse struct {
	Height          float64 `json:"height"`
	Weight          float64 `json:"weight"`
	PredictedHeight float64 `json:"predicted_height"`
	Confidence      float64 `json:"confidence"`
	Error           string  `json:"error,omitempty"`
}

// PredictWeight sends the front and side images along with height to the model service
// and returns the predicted weight
func PredictWeight(frontImgPath, sideImgPath string, height float64) (float64, error) {
	// Load config properly with error handling
	cfg, err := config.LoadConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to load config: %w", err)
	}

	// Get model service URL
	modelServiceURL := cfg.MLServiceURL + "/predict"
	fmt.Printf("Sending prediction request to: %s\n", modelServiceURL)

	// If in DEV_MODE, use mock implementation
	if cfg.MLServiceURL == "" || os.Getenv("DEV_MODE") == "true" {
		fmt.Println("WARNING: Using mock weight prediction instead of ML model")
		weight := (height - 100) * 0.9
		frontInfo, err := os.Stat(frontImgPath)
		if err == nil {
			weight += float64(frontInfo.Size()%10) * 0.1
		}
		sideInfo, err := os.Stat(sideImgPath)
		if err == nil {
			weight += float64(sideInfo.Size()%10) * 0.1
		}
		return weight, nil
	}

	// Create multipart form data
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Open and add front image file
	frontFile, err := os.Open(frontImgPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open front image: %w", err)
	}
	defer frontFile.Close()

	frontFormFile, err := multipartWriter.CreateFormFile("front_image", filepath.Base(frontImgPath))
	if err != nil {
		return 0, fmt.Errorf("failed to create form file for front image: %w", err)
	}
	if _, err = io.Copy(frontFormFile, frontFile); err != nil {
		return 0, fmt.Errorf("failed to copy front image to form: %w", err)
	}

	// Open and add side image file
	sideFile, err := os.Open(sideImgPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open side image: %w", err)
	}
	defer sideFile.Close()

	sideFormFile, err := multipartWriter.CreateFormFile("side_image", filepath.Base(sideImgPath))
	if err != nil {
		return 0, fmt.Errorf("failed to create form file for side image: %w", err)
	}
	if _, err = io.Copy(sideFormFile, sideFile); err != nil {
		return 0, fmt.Errorf("failed to copy side image to form: %w", err)
	}

	// Add height as form field
	heightField, err := multipartWriter.CreateFormField("height")
	if err != nil {
		return 0, fmt.Errorf("failed to create form field for height: %w", err)
	}
	if _, err = heightField.Write([]byte(strconv.FormatFloat(height, 'f', -1, 64))); err != nil {
		return 0, fmt.Errorf("failed to write height to form: %w", err)
	}

	// Close multipart writer
	if err = multipartWriter.Close(); err != nil {
		return 0, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", modelServiceURL, &requestBody)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for multipart form data
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request to model service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response for debugging
	fmt.Printf("Response from ML service (status %d): %s\n", resp.StatusCode, string(body))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("model service returned error status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var modelResponse ModelResponse
	if err := json.Unmarshal(body, &modelResponse); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for error
	if modelResponse.Error != "" {
		return 0, fmt.Errorf("model service error: %s", modelResponse.Error)
	}

	// Store the estimation RESULTS in MongoDB (without storing the actual images)
	if models.DB != nil {
		// Only store metadata and results - not the actual images
		estimation := &models.WeightEstimation{
			ID:        primitive.NewObjectID(),
			Height:    height,
			Weight:    modelResponse.Weight,
			CreatedAt: time.Now(),
			// You can store image paths to temporary files if needed
			// But don't store the actual image data
		}

		if err := models.SaveWeightEstimation(estimation); err != nil {
			fmt.Printf("Failed to save estimation to database: %v\n", err)
			// Continue anyway - don't fail the request
		}
	}

	// Return the weight from the response
	return modelResponse.Weight, nil
}
