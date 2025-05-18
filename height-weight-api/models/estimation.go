package models

import (
	"time"
)

// Estimation represents the height and weight estimation result
type Estimation struct {
	ID        string    `json:"id" bson:"id"`
	ImagePath string    `json:"image_path" bson:"image_path"`
	Height    float64   `json:"height" bson:"height"`     // Height in centimeters
	Weight    float64   `json:"weight" bson:"weight"`     // Weight in kilograms
	Accuracy  float64   `json:"accuracy" bson:"accuracy"` // Estimation accuracy percentage
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// EstimationResult is the response sent to clients
type EstimationResult struct {
	ID        string    `json:"id"`
	Height    float64   `json:"height"`
	Weight    float64   `json:"weight"`
	Accuracy  float64   `json:"accuracy"`
	CreatedAt time.Time `json:"created_at"`
}

// MLServiceRequest is the request sent to the ML service
type MLServiceRequest struct {
	ImagePath string `json:"image_path"`
}

// MLServiceResponse is the response from the ML service
// MLServiceResponse represents the response from the ML service
type MLServiceResponse struct {
	Height     float64 `json:"height"`
	Weight     float64 `json:"weight"`
	Confidence float64 `json:"confidence"`
	Error      string  `json:"error,omitempty"`
}
