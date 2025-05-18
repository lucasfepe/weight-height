package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lucasfepe/height-weight-api/config"
	"github.com/lucasfepe/height-weight-api/db"
	"github.com/lucasfepe/height-weight-api/models"
	"github.com/lucasfepe/height-weight-api/utils"
)

// NewImageUploadHandler creates a handler for image uploads with config
func NewImageUploadHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form with specified max memory
		if err := r.ParseMultipartForm(cfg.MaxFileSize); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request: "+err.Error())
			return
		}

		// Get file from form
		file, fileHeader, err := r.FormFile("image")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Failed to get image: "+err.Error())
			return
		}
		defer file.Close()

		// Validate file size
		if fileHeader.Size > cfg.MaxFileSize {
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("File too large. Max size: %d bytes", cfg.MaxFileSize))
			return
		}

		// Validate file extension
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		validExt := false
		for _, allowedExt := range cfg.AllowedExts {
			if ext == allowedExt {
				validExt = true
				break
			}
		}
		if !validExt {
			utils.RespondWithError(w, http.StatusBadRequest, "Unsupported file format")
			return
		}

		// Generate unique ID and save file
		imageID := uuid.New().String()
		filename := imageID + ext
		filePath := filepath.Join(cfg.UploadDir, filename)

		// Create file
		dst, err := os.Create(filePath)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create file: "+err.Error())
			return
		}
		defer dst.Close()

		// Copy file content
		if _, err = io.Copy(dst, file); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save file: "+err.Error())
			return
		}

		// Reopen file for reading to send to ML service
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to read saved file: "+err.Error())
			return
		}

		// Call ML service for estimation
		result, err := callMLService(fileContent, cfg.MLServiceURL)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process image: "+err.Error())
			return
		}

		// Create and store estimation result
		estimation := models.Estimation{
			ID:        imageID,
			ImagePath: filePath,
			Height:    result.Height,
			Weight:    result.Weight,
			Accuracy:  result.Confidence, // Note: adjusted field name from the ML service
			CreatedAt: time.Now(),
		}

		// Save to MongoDB
		if err := db.SaveEstimation(&estimation); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save estimation: "+err.Error())
			return
		}

		// Return result
		response := models.EstimationResult{
			ID:        estimation.ID,
			Height:    estimation.Height,
			Weight:    estimation.Weight,
			Accuracy:  estimation.Accuracy,
			CreatedAt: estimation.CreatedAt,
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

// callMLService calls the Python ML service for height and weight estimation
func callMLService(imageData []byte, mlServiceURL string) (*models.MLServiceResponse, error) {
	// Create a new multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form file field
	part, err := writer.CreateFormFile("image", "image.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Write the image data to the form
	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("failed to write image to form: %w", err)
	}

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create and send the HTTP request
	req, err := http.NewRequest("POST", mlServiceURL+"/predict", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned error: %s, body: %s", resp.Status, string(respBody))
	}

	// Parse the response
	var result models.MLServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse ML service response: %w", err)
	}

	return &result, nil
}
