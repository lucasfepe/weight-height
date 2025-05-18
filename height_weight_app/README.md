# Height and Weight Estimator App

A Flutter mobile application that captures images and communicates with a Go backend to estimate a person's height and weight using a TensorFlow model.

## Features

- Capture full-body photos using the device camera
- Select existing images from the photo gallery
- Display estimated height and weight based on image analysis
- Communicate with the Go backend API
- Guidelines overlay to help position the subject properly

## Project Structure

```
height_weight_app/
├── lib/
│   ├── main.dart              # Entry point of the application
│   ├── screens/               # App screens
│   │   ├── home_screen.dart   # Main screen with image capture/selection
│   │   └── camera_screen.dart # Custom camera interface
│   ├── widgets/               # Reusable UI components
│   │   └── result_display.dart # Display height/weight results
│   ├── services/              # API and backend services
│   │   └── api_service.dart   # Communication with Go backend
│   ├── models/                # Data models
│   └── utils/                 # Utility functions and constants
│       └── constants.dart     # App-wide constants
├── android/                   # Android-specific configuration
├── ios/                       # iOS-specific configuration
├── assets/                    # App assets (images, etc.)
└── pubspec.yaml               # Flutter dependencies
```

## Setup and Installation

### Prerequisites

- Flutter SDK
- Android Studio / Xcode for device emulation
- Go backend running with the height-weight estimation model

### Installation

1. Clone this repository
2. Run `flutter pub get` to install dependencies
3. Update the API base URL in `lib/utils/constants.dart` to point to your Go backend
4. Connect a device or start an emulator
5. Run `flutter run` to start the app

## Usage

1. Launch the app
2. Choose to take a new photo or select one from your gallery
3. For best results, ensure the photo:
   - Shows the full body
   - Is taken from a consistent distance
   - Has the subject standing straight
4. Press the "ANALYZE" button to send the image to the backend
5. View the estimated height and weight results

## Communication with Go Backend

The app communicates with the Go backend through a REST API:
- Endpoint: `POST /estimate`
- Request: Multipart form with image file
- Response: JSON with estimated height and weight values

## Permissions

The app requires the following permissions:
- Camera access for taking photos
- Photo library access for selecting images
- Internet access for API communication 