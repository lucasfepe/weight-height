package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lucasfepe/height-weight-api/db"
	"github.com/lucasfepe/height-weight-api/models"
	"github.com/lucasfepe/height-weight-api/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetEstimationHandler handles requests to get estimation results by ID
func GetEstimationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["imageID"]

	if imageID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing image ID")
		return
	}

	// Fetch estimation from MongoDB
	estimation, err := db.GetEstimationByID(imageID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Estimation not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve estimation: "+err.Error())
		}
		return
	}

	// Create response
	result := models.EstimationResult{
		ID:        estimation.ID,
		Height:    estimation.Height,
		Weight:    estimation.Weight,
		Accuracy:  estimation.Accuracy,
		CreatedAt: estimation.CreatedAt,
	}

	utils.RespondWithJSON(w, http.StatusOK, result)
}

// ListEstimationsHandler returns a list of estimations with pagination
func ListEstimationsHandler(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0

	// Parse query parameters for pagination
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if _, err := fmt.Sscanf(limitParam, "%d", &limit); err != nil {
			limit = 10
		}
	}

	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if _, err := fmt.Sscanf(offsetParam, "%d", &offset); err != nil {
			offset = 0
		}
	}

	// Get estimations from database
	estimations, err := db.ListEstimations(limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve estimations: "+err.Error())
		return
	}

	// Convert to response format
	var results []models.EstimationResult
	for _, est := range estimations {
		results = append(results, models.EstimationResult{
			ID:        est.ID,
			Height:    est.Height,
			Weight:    est.Weight,
			Accuracy:  est.Accuracy,
			CreatedAt: est.CreatedAt,
		})
	}

	utils.RespondWithJSON(w, http.StatusOK, results)
}

// DeleteEstimationHandler deletes an estimation by ID
func DeleteEstimationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["imageID"]

	if imageID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing image ID")
		return
	}

	// First get the estimation to check if it exists and to get the image path
	estimation, err := db.GetEstimationByID(imageID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Estimation not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve estimation: "+err.Error())
		}
		return
	}

	// Delete from database
	if err := db.DeleteEstimation(imageID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete estimation: "+err.Error())
		return
	}

	// Delete the image file
	if err := os.Remove(estimation.ImagePath); err != nil {
		// Just log this error, don't fail the request
		log.Printf("Warning: Failed to delete image file %s: %v", estimation.ImagePath, err)
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Estimation deleted successfully"})
}
