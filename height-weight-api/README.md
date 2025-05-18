# Height and Weight Estimation API

A Go backend API server for a height and weight estimation application. The API allows clients to upload images and receive height and weight estimations from a Python ML service.

## Project Structure

```
height-weight-api/
├── api/
│   └── router.go         # Router setup
├── config/
│   └── config.go         # Application configuration
├── handlers/
│   ├── estimation.go     # Estimation retrieval handler
│   ├── health.go         # Health check handler
│   └── upload.go         # Image upload handler
├── models/
│   └── estimation.go     # Data models
├── utils/
│   └── response.go       # HTTP response utilities
├── main.go               # Application entry point
├── go.mod                # Go module dependencies
└── README.md             # This file
```

## Prerequisites

- Go 1.21 or later
- Python 3.8+ (for the ML service)

## Configuration

The application can be configured using environment variables:

- `PORT`: Server port (default: 8080)
- `ML_SERVICE_URL`: URL of the Python ML service (default: http://localhost:5000)
- `UPLOAD_DIR`: Directory to store uploaded images (default: ./uploads)

## Getting Started

1. Clone the repository
2. Install dependencies:

```bash
go mod download
```

3. Run the server:

```bash
go run main.go
```

## API Endpoints

### Health Check

```
GET /api/health
```

Response:
```json
{
  "status": "OK"
}
```

### Upload Image

```
POST /api/upload
```

Request:
- Form data with "image" field containing the image file

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "height": 175.5,
  "weight": 70.2,
  "accuracy": 0.92,
  "created_at": "2023-11-01T12:34:56Z"
}
```

### Get Estimation Results

```
GET /api/estimate/{imageID}
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "height": 175.5,
  "weight": 70.2,
  "accuracy": 0.92,
  "created_at": "2023-11-01T12:34:56Z"
}
```

## ML Service Integration

The API server expects the ML service to expose an endpoint:

```
POST /predict
```

Request:
```json
{
  "image_path": "/path/to/image.jpg"
}
```

Response:
```json
{
  "height": 175.5,
  "weight": 70.2,
  "accuracy": 0.92
}
``` 