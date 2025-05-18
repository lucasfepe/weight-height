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
	"github.com/lucasfepe/height-weight-api/utils"
)

// Response represents the standard API response format
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// EstimateWeight handles the weight estimation based on front image, side image, and height
func EstimateWeight(w http.ResponseWriter, r *http.Request) {
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

	height, err := strconv.ParseFloat(heightStr, 64)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid height value: "+err.Error())
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
	if err := os.MkdirAll("uploads", 0755); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create uploads directory: "+err.Error())
		return
	}

	// Create timestamp for unique filenames
	timestamp := time.Now().UnixNano()

	// Save front image
	frontFilename := fmt.Sprintf("%d_%s", timestamp, frontHeader.Filename)
	frontFilepath := filepath.Join("uploads", frontFilename)
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
	sideFilename := fmt.Sprintf("%d_%s", timestamp, sideHeader.Filename)
	sideFilepath := filepath.Join("uploads", sideFilename)
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

	// Process images with the TensorFlow model
	weight, err := utils.PredictWeight(frontFilepath, sideFilepath, height)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to predict weight: "+err.Error())
		return
	}

	// Create a record of the estimation
	estimation := &models.WeightEstimation{
		Height:       height,
		Weight:       weight,
		FrontImgPath: frontFilepath,
		SideImgPath:  sideFilepath,
		CreatedAt:    time.Now(),
	}

	// Save the estimation record to database (if db is set up)
	if models.DB != nil {
		if err := models.SaveWeightEstimation(estimation); err != nil {
			// Log the error but don't fail the request
			fmt.Printf("Failed to save estimation to database: %v\n", err)
		}
	}

	// Return the estimated weight
	response := Response{
		Success: true,
		Data: map[string]float64{
			"weight": weight,
		},
		Message: "Weight estimated successfully",
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper function to send error responses
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	response := Response{
		Success: false,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}
