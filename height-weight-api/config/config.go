package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	MLServiceURL    string
	MaxFileSize     int64
	AllowedExts     []string
	UploadDir       string
	MongoURI        string
	MongoDB         string
	MongoCollection string
	MongoTimeout    time.Duration
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() (*Config, error) {
	mlServiceURL := os.Getenv("ML_SERVICE_URL")
	if mlServiceURL == "" {
		mlServiceURL = "http://localhost:5000" // Default ML service URL
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// MongoDB configuration - use environment variables for credentials
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb+srv://user:xwt3IRyuDOaN5MxP@project.fnqowfy.mongodb.net/Transit?retryWrites=true&w=majority&appName=Project"
	}

	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "height_weight_app"
	}

	mongoCollection := os.Getenv("MONGO_COLLECTION")
	if mongoCollection == "" {
		mongoCollection = "estimations"
	}

	mongoTimeoutSec := 10
	if timeoutStr := os.Getenv("MONGO_TIMEOUT_SEC"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			mongoTimeoutSec = timeout
		}
	}

	// Parse max file size from environment or use default
	maxFileSizeMB := 10 // Default 10MB
	if sizeStr := os.Getenv("MAX_FILE_SIZE_MB"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			maxFileSizeMB = size
		}
	}

	// Ensure upload directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.MkdirAll(uploadDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	return &Config{
		MLServiceURL:    mlServiceURL,
		MaxFileSize:     int64(maxFileSizeMB) * 1024 * 1024,
		AllowedExts:     []string{".jpg", ".jpeg", ".png"},
		UploadDir:       uploadDir,
		MongoURI:        mongoURI,
		MongoDB:         mongoDB,
		MongoCollection: mongoCollection,
		MongoTimeout:    time.Duration(mongoTimeoutSec) * time.Second,
	}, nil
}
