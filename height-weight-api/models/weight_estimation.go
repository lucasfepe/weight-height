package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

// WeightEstimation represents a weight estimation record
type WeightEstimation struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Height       float64            `bson:"height" json:"height"`
	Weight       float64            `bson:"weight" json:"weight"`
	FrontImgPath string             `bson:"front_img_path" json:"front_img_path"`
	SideImgPath  string             `bson:"side_img_path" json:"side_img_path"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

// SaveWeightEstimation saves the weight estimation to the database
func SaveWeightEstimation(estimation *WeightEstimation) error {
	// Set created_at timestamp if not set
	if estimation.CreatedAt.IsZero() {
		estimation.CreatedAt = time.Now()
	}

	// Set a new ID if not provided
	if estimation.ID.IsZero() {
		estimation.ID = primitive.NewObjectID()
	}

	// Get the collection
	collection := DB.Collection("weight_estimations")

	// Insert the document
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, estimation)
	return err
}

// GetWeightEstimations retrieves weight estimations from the database
func GetWeightEstimations(limit int64) ([]*WeightEstimation, error) {
	// Get the collection
	collection := DB.Collection("weight_estimations")

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
	var results []*WeightEstimation
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
