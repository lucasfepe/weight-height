# service.py
import tensorflow as tf
import numpy as np
import cv2
import base64
import io
from flask import Flask, request, jsonify
from height_weight_estimator import preprocess_image
import os

app = Flask(__name__)

# Global variable to store the loaded model
model = None

def load_model_if_needed(model_path):
    """Load the model if it's not already loaded"""
    global model
    if model is None:
        model = tf.keras.models.load_model(model_path)
    return model

def process_image_data(image_data):
    """Process image data from base64 string"""
    # Decode base64 string to bytes
    image_bytes = base64.b64decode(image_data)
    
    # Convert bytes to numpy array
    nparr = np.frombuffer(image_bytes, np.uint8)
    
    # Decode image
    img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
    
    # Convert BGR to RGB
    img = cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
    
    # Preprocess the image
    img = preprocess_image(img)
    
    # Add batch dimension
    img = np.expand_dims(img, axis=0)
    
    return img

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({"status": "healthy"})

@app.route('/predict', methods=['POST'])
def predict():
    """Predict height and weight from an image"""
    # Check if request has the required data
    if not request.json or 'image' not in request.json:
        return jsonify({"error": "Missing image data"}), 400
    
    # Get model path from request or use default
    model_path = request.json.get('model_path', 'model.h5')
    
    try:
        # Load the model
        current_model = load_model_if_needed(model_path)
        
        # Process the image
        image_data = request.json['image']
        img = process_image_data(image_data)
        
        # Make prediction
        predictions = current_model.predict(img)
        
        # Extract height and weight predictions
        if isinstance(predictions, list):
            # If model returns a list of outputs
            predicted_height = float(predictions[0][0][0])
            predicted_weight = float(predictions[1][0][0])
        else:
            # If model returns a single array with both values
            predicted_height = float(predictions[0][0])
            predicted_weight = float(predictions[0][1])
        
        # Return the results
        return jsonify({
            "height": predicted_height,
            "weight": predicted_weight
        })
    
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    # Default port is 5000
    port = int(os.environ.get('PORT', 5000))
    
    # Run the Flask app
    app.run(host='0.0.0.0', port=port, debug=True)
    print(f"Running prediction service on http://0.0.0.0:{port}")