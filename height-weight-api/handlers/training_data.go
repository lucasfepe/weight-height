package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/lucasfepe/height-weight-api/models"
)

// SaveTrainingData handles saving training data (images + actual weight + height)
func SaveTrainingData(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse the multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		sendErrorResponse(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	// Get height from form
	heightStr := r.FormValue("height")
	if heightStr == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Height is required")
		return
	}

	// Get actual weight from form
	actualWeightStr := r.FormValue("actual_weight")
	if actualWeightStr == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Actual weight is required")
		return
	}

	// Parse values
	height, err := strconv.ParseFloat(heightStr, 64)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid height value: "+err.Error())
		return
	}

	actualWeight, err := strconv.ParseFloat(actualWeightStr, 64)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid weight value: "+err.Error())
		return
	}

	// Get front image from form
	frontFile, frontHeader, err := r.FormFile("front_image")
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Front image is required: "+err.Error())
		return
	}
	defer frontFile.Close()

	// Get side image from form
	sideFile, sideHeader, err := r.FormFile("side_image")
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Side image is required: "+err.Error())
		return
	}
	defer sideFile.Close()

	// Create uploads directory if it doesn't exist
	trainingDir := filepath.Join("uploads", "training")
	if err := os.MkdirAll(trainingDir, 0755); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create uploads directory: "+err.Error())
		return
	}

	// Create timestamp for unique filenames
	timestamp := time.Now().UnixNano()

	// Save front image
	frontFilename := fmt.Sprintf("train_%d_%s", timestamp, frontHeader.Filename)
	frontFilepath := filepath.Join(trainingDir, frontFilename)
	frontDst, err := os.Create(frontFilepath)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to save front image: "+err.Error())
		return
	}
	defer frontDst.Close()

	if _, err = io.Copy(frontDst, frontFile); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to save front image data: "+err.Error())
		return
	}

	// Save side image
	sideFilename := fmt.Sprintf("train_%d_%s", timestamp, sideHeader.Filename)
	sideFilepath := filepath.Join(trainingDir, sideFilename)
	sideDst, err := os.Create(sideFilepath)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to save side image: "+err.Error())
		return
	}
	defer sideDst.Close()

	if _, err = io.Copy(sideDst, sideFile); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to save side image data: "+err.Error())
		return
	}

	// Create a training data record
	trainingData := &models.TrainingData{
		Height:       height,
		ActualWeight: actualWeight,
		FrontImgPath: frontFilepath,
		SideImgPath:  sideFilepath,
		CreatedAt:    time.Now(),
	}

	// Save the training data record to database
	if models.DB != nil {
		if err := models.SaveTrainingData(trainingData); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to save training data to database: "+err.Error())
			return
		}
	}

	// Return success response
	response := Response{
		Success: true,
		Data: map[string]interface{}{
			"id":            trainingData.ID.Hex(),
			"height":        trainingData.Height,
			"actual_weight": trainingData.ActualWeight,
			"created_at":    trainingData.CreatedAt,
		},
		Message: "Training data saved successfully",
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTrainingData returns a list of training data records
func GetTrainingData(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	if models.DB == nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get limit parameter (optional)
	limitStr := r.URL.Query().Get("limit")
	var limit int64 = 50 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get training data from database
	trainingData, err := models.GetTrainingData(limit)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch training data: "+err.Error())
		return
	}

	// Return success response
	response := Response{
		Success: true,
		Data:    trainingData,
		Message: fmt.Sprintf("Retrieved %d training data records", len(trainingData)),
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ExportTrainingData exports all training data for model training
func ExportTrainingData(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	if models.DB == nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get all training data
	trainingData, err := models.ExportTrainingData()
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch training data: "+err.Error())
		return
	}

	// Format data for export
	type ExportData struct {
		FrontImgPath string  `json:"front_image_path"`
		SideImgPath  string  `json:"side_image_path"`
		Height       float64 `json:"height"`
		ActualWeight float64 `json:"actual_weight"`
	}

	exportData := make([]ExportData, len(trainingData))
	for i, td := range trainingData {
		exportData[i] = ExportData{
			FrontImgPath: td.FrontImgPath,
			SideImgPath:  td.SideImgPath,
			Height:       td.Height,
			ActualWeight: td.ActualWeight,
		}
	}

	// Return success response
	response := Response{
		Success: true,
		Data:    exportData,
		Message: fmt.Sprintf("Exported %d training data records", len(trainingData)),
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
