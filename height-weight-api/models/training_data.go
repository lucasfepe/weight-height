package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TrainingData represents a record for training data
type TrainingData struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Height       float64            `bson:"height" json:"height"`
	ActualWeight float64            `bson:"actual_weight" json:"actual_weight"`
	FrontImgPath string             `bson:"front_img_path" json:"front_img_path"`
	SideImgPath  string             `bson:"side_img_path" json:"side_img_path"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

// SaveTrainingData saves the training data to the database
func SaveTrainingData(data *TrainingData) error {
	// Set created_at timestamp if not set
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}

	// Set a new ID if not provided
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
	}

	// Get the collection
	collection := DB.Collection("training_data")

	// Insert the document
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, data)
	return err
}

// GetTrainingData retrieves training data from the database
func GetTrainingData(limit int64) ([]*TrainingData, error) {
	// Get the collection
	collection := DB.Collection("training_data")

	// Set up the query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by created_at desc
	if limit > 0 {
		findOptions.SetLimit(limit)
	}

	// Execute the query
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the results
	var results []*TrainingData
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// ExportTrainingData returns all training data formatted for model training
func ExportTrainingData() ([]*TrainingData, error) {
	// Get all training data without limit
	return GetTrainingData(0)
}
