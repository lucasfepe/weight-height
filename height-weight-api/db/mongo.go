package db

import (
	"context"
	"time"

	"github.com/lucasfepe/height-weight-api/config"
	"github.com/lucasfepe/height-weight-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var collection *mongo.Collection

// InitMongoDB initializes the MongoDB connection
// InitMongoDB initializes the MongoDB connection
func InitMongoDB(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.MongoTimeout)
	defer cancel()

	var err error
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	// Get a handle to the estimations collection
	collection = client.Database(cfg.MongoDB).Collection(cfg.MongoCollection)

	// Initialize the models.DB variable for use in weight_estimation.go
	models.DB = client.Database(cfg.MongoDB)

	// Create indexes for faster lookups
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"id", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// CloseMongoDB closes the MongoDB connection
func CloseMongoDB() error {
	if client == nil {
		return nil
	}
	return client.Disconnect(context.Background())
}

// SaveEstimation saves an estimation to MongoDB
func SaveEstimation(estimation *models.Estimation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, estimation)
	return err
}

// GetEstimationByID retrieves an estimation by ID
func GetEstimationByID(id string) (*models.Estimation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var estimation models.Estimation
	filter := bson.M{"id": id}
	err := collection.FindOne(ctx, filter).Decode(&estimation)

	if err != nil {
		return nil, err
	}

	return &estimation, nil
}

// ListEstimations retrieves a list of estimations with pagination
func ListEstimations(limit, offset int) ([]models.Estimation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))
	findOptions.SetSort(bson.D{{"created_at", -1}}) // Sort by newest first

	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var estimations []models.Estimation
	if err := cursor.All(ctx, &estimations); err != nil {
		return nil, err
	}

	return estimations, nil
}

// DeleteEstimation deletes an estimation by ID
func DeleteEstimation(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"id": id}
	_, err := collection.DeleteOne(ctx, filter)
	return err
}
