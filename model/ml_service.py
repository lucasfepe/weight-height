from flask import Flask, request, jsonify
import tensorflow as tf
import numpy as np
import cv2
import os
import traceback
import sys
from height_weight_estimator import preprocess_image
import io
import logging

# Configure logging
logging.basicConfig(level=logging.DEBUG, 
                    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
                    handlers=[logging.StreamHandler()])
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Load model at startup to avoid reloading for each request
MODEL_PATH = os.environ.get('MODEL_PATH', 'models/height_weight_model.h5')
model = None

def load_model():
    global model
    if model is None:
        try:
            # Try both file extensions
            keras_path = 'models/height_weight_model.keras'
            h5_path = 'models/height_weight_model.h5'
            
            if os.path.exists(keras_path):
                actual_path = keras_path
                logger.info(f"Found model file with .keras extension: {actual_path}")
            elif os.path.exists(h5_path):
                actual_path = h5_path
                logger.info(f"Found model file with .h5 extension: {actual_path}")
            else:
                error_msg = f"Model file not found at {keras_path} or {h5_path}"
                logger.error(error_msg)
                raise FileNotFoundError(error_msg)
            
            logger.info(f"Loading model from: {actual_path}")
            
            # Use corrected API for newer TensorFlow versions
            # Define custom objects for newer TensorFlow
            try:
                # First try modern Keras API paths
                custom_objects = {
                    'mean_squared_error': tf.keras.losses.MeanSquaredError(),
                    'mean_absolute_error': tf.keras.metrics.MeanAbsoluteError()
                }
                model = tf.keras.models.load_model(actual_path, custom_objects=custom_objects)
            except Exception as e:
                logger.warning(f"Failed to load with first method: {e}")
                # Try alternate API paths
                model = tf.keras.models.load_model(actual_path, compile=False)
                logger.info("Loaded model with compile=False")
            
            logger.info("Model loaded successfully")
            logger.debug(f"Model summary: {model.summary()}")
            
            # Add debug for model inputs
            logger.info("Model input layers information:")
            for i, layer in enumerate(model.inputs):
                logger.info(f"  Input {i}: name={layer.name}, shape={layer.shape}")
                
        except Exception as e:
            logger.error(f"Error loading model: {str(e)}")
            logger.error(traceback.format_exc())
            raise
    return model

def preprocess_uploaded_images(front_image_bytes, side_image_bytes):
    """
    Process front and side images uploaded as bytes.
    """
    try:
        logger.info(f"Preprocessing front image, bytes length: {len(front_image_bytes)}")
        logger.info(f"Preprocessing side image, bytes length: {len(side_image_bytes)}")
        
        # Process front image
        nparr_front = np.frombuffer(front_image_bytes, np.uint8)
        logger.debug(f"Front NumPy array created, shape: {nparr_front.shape}")
        
        img_front = cv2.imdecode(nparr_front, cv2.IMREAD_COLOR)
        if img_front is None:
            logger.error("Failed to decode front image with OpenCV")
            raise ValueError("Invalid front image format or corrupted front image data")
        
        logger.debug(f"Front image decoded, shape: {img_front.shape}")
        
        # Convert BGR to RGB for front image
        img_front = cv2.cvtColor(img_front, cv2.COLOR_BGR2RGB)
        logger.debug("Front image BGR to RGB conversion complete")
        
        # Preprocess the front image
        img_front = preprocess_image(img_front)
        logger.debug(f"Front image preprocessed, shape: {img_front.shape}")
        
        # Process side image
        nparr_side = np.frombuffer(side_image_bytes, np.uint8)
        logger.debug(f"Side NumPy array created, shape: {nparr_side.shape}")
        
        img_side = cv2.imdecode(nparr_side, cv2.IMREAD_COLOR)
        if img_side is None:
            logger.error("Failed to decode side image with OpenCV")
            raise ValueError("Invalid side image format or corrupted side image data")
        
        logger.debug(f"Side image decoded, shape: {img_side.shape}")
        
        # Convert BGR to RGB for side image
        img_side = cv2.cvtColor(img_side, cv2.COLOR_BGR2RGB)
        logger.debug("Side image BGR to RGB conversion complete")
        
        # Preprocess the side image
        img_side = preprocess_image(img_side)
        logger.debug(f"Side image preprocessed, shape: {img_side.shape}")
        
        # Add batch dimension to both images
        img_front = np.expand_dims(img_front, axis=0)
        img_side = np.expand_dims(img_side, axis=0)
        
        logger.debug(f"Front image with batch dimension, shape: {img_front.shape}")
        logger.debug(f"Side image with batch dimension, shape: {img_side.shape}")
        
        return img_front, img_side
    except Exception as e:
        logger.error(f"Error in image preprocessing: {str(e)}")
        logger.error(traceback.format_exc())
        raise

def adjust_weight_prediction(predicted_weight, user_height, predicted_height):
    """
    Adjust weight prediction when user height is provided.
    Uses the principle that weight roughly scales with the square of height (BMI principle).
    """
    try:
        if user_height <= 0:
            logger.warning("Invalid user height provided, using model's height prediction")
            return predicted_weight
            
        # Calculate adjustment factor based on height difference
        # Using squared ratio to account for BMI-like scaling
        height_ratio = user_height / predicted_height
        adjustment_factor = height_ratio ** 2
        
        # Apply adjustment to weight prediction
        adjusted_weight = predicted_weight * adjustment_factor
        
        logger.info(f"Weight adjusted from {predicted_weight} to {adjusted_weight} based on user height {user_height}cm")
        return adjusted_weight
    except Exception as e:
        logger.error(f"Error adjusting weight: {str(e)}")
        # Return original prediction if adjustment fails
        return predicted_weight

@app.route('/predict', methods=['POST'])
def predict():
    logger.info("Received request to /predict endpoint")
    
    # Check if both images were uploaded
    if 'front_image' not in request.files or 'side_image' not in request.files:
        logger.warning("Missing front or side image in request")
        return jsonify({'error': 'Both front and side images are required'}), 400
    
    front_file = request.files['front_image']
    side_file = request.files['side_image']
    
    # Check for user-provided height (optional)
    user_height = None
    if 'height' in request.form:
        try:
            user_height = float(request.form['height'])
            logger.info(f"User provided height: {user_height}cm")
            if user_height <= 0 or user_height > 300:  # Basic validation
                logger.warning(f"Suspicious height value provided: {user_height}cm")
                return jsonify({'error': 'Invalid height value. Please provide a realistic height in cm.'}), 400
        except ValueError:
            logger.warning(f"Invalid height format provided: {request.form['height']}")
            return jsonify({'error': 'Height must be a valid number in cm'}), 400
    
    logger.info(f"Received front file: {front_file.filename}, content type: {front_file.content_type}")
    logger.info(f"Received side file: {side_file.filename}, content type: {side_file.content_type}")
    
    if front_file.filename == '' or side_file.filename == '':
        logger.warning("Empty filename received")
        return jsonify({'error': 'Both front and side images must be selected'}), 400
    
    try:
        # Read image bytes
        logger.info("Reading front and side image files")
        front_image_bytes = front_file.read()
        side_image_bytes = side_file.read()
        logger.info(f"Front image read, size: {len(front_image_bytes)} bytes")
        logger.info(f"Side image read, size: {len(side_image_bytes)} bytes")
        
        # Preprocess both images
        logger.info("Starting image preprocessing for both images")
        img_front, img_side = preprocess_uploaded_images(front_image_bytes, side_image_bytes)
        logger.info("Both images preprocessed successfully")
        
        # Get the model
        logger.info("Loading model")
        loaded_model = load_model()
        logger.info("Model loaded successfully for prediction")
        
        # Debug the model's input layers
        logger.info("Model input layer names:")
        for i, layer in enumerate(loaded_model.inputs):
            logger.info(f"  Input {i}: {layer.name}")
        
        # Try different input combinations based on model structure
        try:
            # Get all input layer names
            input_layer_names = [layer.name for layer in loaded_model.inputs]
            logger.info(f"Available input layer names: {input_layer_names}")
            
            # Check if the model has a single input layer named "input_layer"
            if "input_layer" in input_layer_names and len(input_layer_names) == 1:
                logger.info("Using single input_layer with front image only")
                predictions = loaded_model.predict({'input_layer': img_front})
            # Check if we have front_input and side_input
            elif "front_input" in input_layer_names and "side_input" in input_layer_names:
                logger.info("Using front_input and side_input")
                predictions = loaded_model.predict({
                    'front_input': img_front, 
                    'side_input': img_side
                })
            # If there's exactly one input layer, use that with the front image
            elif len(input_layer_names) == 1:
                input_name = input_layer_names[0]
                logger.info(f"Using single input layer '{input_name}' with front image")
                predictions = loaded_model.predict({input_name: img_front})
            # Try to concatenate images for a single input
            elif len(input_layer_names) == 1:
                input_name = input_layer_names[0]
                logger.info(f"Trying to concatenate images for input '{input_name}'")
                # Concatenate along the channel dimension
                combined_input = np.concatenate([img_front, img_side], axis=3)
                logger.info(f"Combined input shape: {combined_input.shape}")
                predictions = loaded_model.predict({input_name: combined_input})
            # Fallback - try using functional model direct predict
            else:
                logger.info("Using direct model.predict call without dictionary")
                if len(loaded_model.inputs) == 1:
                    predictions = loaded_model.predict(img_front)
                elif len(loaded_model.inputs) == 2:
                    predictions = loaded_model.predict([img_front, img_side])
                else:
                    raise ValueError(f"Unsupported number of inputs: {len(loaded_model.inputs)}")
        except Exception as predict_error:
            logger.error(f"Error during prediction: {predict_error}")
            logger.error(traceback.format_exc())
            
            # Last resort - try with just one image to input_layer
            logger.info("Last resort attempt: Using 'input_layer' with front image")
            try:
                predictions = loaded_model.predict({'input_layer': img_front})
            except Exception as e:
                logger.error(f"Last resort failed too: {e}")
                raise predict_error  # Re-raise the original error
        
        logger.info(f"Prediction made, result type: {type(predictions)}")
        logger.debug(f"Raw predictions: {predictions}")
        
        # Extract height and weight predictions
        logger.info("Extracting height and weight from predictions")
        try:
            # Try list format first
            if isinstance(predictions, list):
                logger.debug("Predictions are in list format")
                predicted_height = float(predictions[0][0][0])
                predicted_weight = float(predictions[1][0][0])
                logger.debug(f"Extracted height: {predicted_height}, weight: {predicted_weight}")
            else:
                # Try dictionary format
                logger.debug("Attempting dictionary format")
                predicted_height = float(predictions['height'][0][0])
                predicted_weight = float(predictions['weight'][0][0])
                logger.debug(f"Extracted height: {predicted_height}, weight: {predicted_weight}")
        except (IndexError, KeyError, TypeError) as e:
            logger.warning(f"Error extracting predictions using standard formats: {str(e)}")
            logger.warning("Trying fallback extraction method")
            
            # Fallback option if the above formats don't work
            if isinstance(predictions, list):
                logger.debug("Using fallback list format")
                predicted_height = float(predictions[0][0])
                predicted_weight = float(predictions[1][0])
            else:
                # Last resort - try flattening any nested structure
                logger.debug("Using last resort flattening method")
                predictions_flat = tf.nest.flatten(predictions)
                logger.debug(f"Flattened predictions: {predictions_flat}")
                
                if len(predictions_flat) >= 2:
                    predicted_height = float(predictions_flat[0])
                    predicted_weight = float(predictions_flat[1])
                else:
                    # If we only have one output, use a simple BMI-based estimate
                    value = float(predictions_flat[0])
                    # Determine if the value is more likely to be height or weight
                    if 100 <= value <= 220:  # Likely height in cm
                        predicted_height = value
                        predicted_weight = (value - 100) * 0.9  # Simple BMI formula
                    else:  # Likely weight in kg
                        predicted_weight = value
                        predicted_height = 170  # Default average height
            
            logger.debug(f"Fallback extraction - height: {predicted_height}, weight: {predicted_weight}")
        
        # If user provided height, adjust the weight prediction
        final_height = user_height if user_height is not None else predicted_height
        final_weight = predicted_weight
        
        if user_height is not None:
            logger.info(f"Adjusting weight based on user-provided height of {user_height}cm")
            final_weight = adjust_weight_prediction(predicted_weight, user_height, predicted_height)
        
        # Increase confidence if user provided their height
        confidence = 0.92 if user_height is not None else 0.90
        
        # Return predictions as JSON
        logger.info(f"Returning prediction - height: {final_height}, weight: {final_weight}, confidence: {confidence}")
        return jsonify({
            'height': final_height,
            'weight': final_weight,
            'predicted_height': predicted_height,  # Include model's height prediction for reference
            'confidence': confidence
        })
    
    except Exception as e:
        logger.error(f"Error in prediction process: {str(e)}")
        logger.error(traceback.format_exc())
        return jsonify({'error': str(e)}), 500

@app.route('/', methods=['GET'])
def health_check():
    logger.info("Health check endpoint called")
    return jsonify({'status': 'Service is running'})

# Add a compatibility endpoint for single image prediction (for backward compatibility)
# Add a compatibility endpoint for single image prediction (for backward compatibility)
@app.route('/predict_single', methods=['POST'])
def predict_single():
    logger.info("Received request to /predict_single endpoint (legacy)")
    
    # Check if image was uploaded
    if 'image' not in request.files:
        logger.warning("No image found in request")
        return jsonify({'error': 'No image uploaded'}), 400
    
    file = request.files['image']
    
    # Check for user-provided height (optional)
    user_height = None
    if 'height' in request.form:
        try:
            user_height = float(request.form['height'])
            logger.info(f"User provided height: {user_height}cm")
            if user_height <= 0 or user_height > 300:  # Basic validation
                logger.warning(f"Suspicious height value provided: {user_height}cm")
                return jsonify({'error': 'Invalid height value. Please provide a realistic height in cm.'}), 400
        except ValueError:
            logger.warning(f"Invalid height format provided: {request.form['height']}")
            return jsonify({'error': 'Height must be a valid number in cm'}), 400
    
    logger.info(f"Received file: {file.filename}, content type: {file.content_type}")
    
    if file.filename == '':
        logger.warning("Empty filename received")
        return jsonify({'error': 'No selected file'}), 400
    
    try:
        # Read image bytes
        logger.info("Reading image file")
        image_bytes = file.read()
        logger.info(f"Image read, size: {len(image_bytes)} bytes")
        
        # For compatibility with single image, we duplicate the same image as front and side
        logger.warning("Using the same image for both front and side views - this is suboptimal")
        img_front, img_side = preprocess_uploaded_images(image_bytes, image_bytes)
        
        # Get the model
        logger.info("Loading model")
        loaded_model = load_model()
        logger.info("Model loaded successfully for prediction")
        
        # Debug the model's input layers
        logger.info("Model input layer names:")
        for i, layer in enumerate(loaded_model.inputs):
            logger.info(f"  Input {i}: {layer.name}")
        
        # Try different input combinations based on model structure
        try:
            # Get all input layer names
            input_layer_names = [layer.name for layer in loaded_model.inputs]
            logger.info(f"Available input layer names: {input_layer_names}")
            
            # Check if the model has a single input layer named "input_layer"
            if "input_layer" in input_layer_names and len(input_layer_names) == 1:
                logger.info("Using single input_layer with image")
                predictions = loaded_model.predict({'input_layer': img_front})
            # Check if we have front_input and side_input
            elif "front_input" in input_layer_names and "side_input" in input_layer_names:
                logger.info("Using front_input and side_input (with duplicated image)")
                predictions = loaded_model.predict({
                    'front_input': img_front, 
                    'side_input': img_side
                })
            # If there's exactly one input layer, use that with the image
            elif len(input_layer_names) == 1:
                input_name = input_layer_names[0]
                logger.info(f"Using single input layer '{input_name}' with image")
                predictions = loaded_model.predict({input_name: img_front})
            # Fallback - try using functional model direct predict
            else:
                logger.info("Using direct model.predict call without dictionary")
                if len(loaded_model.inputs) == 1:
                    predictions = loaded_model.predict(img_front)
                elif len(loaded_model.inputs) == 2:
                    predictions = loaded_model.predict([img_front, img_side])
                else:
                    raise ValueError(f"Unsupported number of inputs: {len(loaded_model.inputs)}")
        except Exception as predict_error:
            logger.error(f"Error during prediction: {predict_error}")
            logger.error(traceback.format_exc())
            
            # Last resort - try with just one image to input_layer
            logger.info("Last resort attempt: Using 'input_layer' with image")
            try:
                predictions = loaded_model.predict({'input_layer': img_front})
            except Exception as e:
                logger.error(f"Last resort failed too: {e}")
                raise predict_error  # Re-raise the original error
        
        # Extract height and weight predictions (same logic as in predict endpoint)
        logger.info("Extracting height and weight from predictions")
        try:
            if isinstance(predictions, list):
                predicted_height = float(predictions[0][0][0])
                predicted_weight = float(predictions[1][0][0])
            else:
                predicted_height = float(predictions['height'][0][0])
                predicted_weight = float(predictions['weight'][0][0])
        except (IndexError, KeyError, TypeError) as e:
            logger.warning(f"Error extracting predictions: {str(e)}")
            
            # Fallback extraction methods
            if isinstance(predictions, list):
                predicted_height = float(predictions[0][0])
                predicted_weight = float(predictions[1][0])
            else:
                # Last resort - try flattening any nested structure
                predictions_flat = tf.nest.flatten(predictions)
                logger.debug(f"Flattened predictions: {predictions_flat}")
                
                if len(predictions_flat) >= 2:
                    predicted_height = float(predictions_flat[0])
                    predicted_weight = float(predictions_flat[1])
                else:
                    # If we only have one output, use a simple BMI-based estimate
                    value = float(predictions_flat[0])
                    # Determine if the value is more likely to be height or weight
                    if 100 <= value <= 220:  # Likely height in cm
                        predicted_height = value
                        predicted_weight = (value - 100) * 0.9  # Simple BMI formula
                    else:  # Likely weight in kg
                        predicted_weight = value
                        predicted_height = 170  # Default average height
        
        # If user provided height, adjust the weight prediction
        final_height = user_height if user_height is not None else predicted_height
        final_weight = predicted_weight
        
        if user_height is not None:
            logger.info(f"Adjusting weight based on user-provided height of {user_height}cm")
            final_weight = adjust_weight_prediction(predicted_weight, user_height, predicted_height)
        
        # Calculate confidence - lower for single image, but higher if height is provided
        confidence = 0.80 if user_height is not None else 0.75
        
        # Return predictions as JSON
        logger.info(f"Returning single-image prediction - height: {final_height}, weight: {final_weight}, confidence: {confidence}")
        return jsonify({
            'height': final_height,
            'weight': final_weight,
            'predicted_height': predicted_height,  # Include model's height prediction for reference
            'confidence': confidence
        })
    
    except Exception as e:
        logger.error(f"Error in single-image prediction process: {str(e)}")
        logger.error(traceback.format_exc())
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    logger.info("ML Service starting...")
    try:
        # Verify model exists before starting the server
        logger.info("Attempting to pre-load model before starting server")
        load_model()
    except Exception as e:
        logger.error(f"Failed to pre-load model: {str(e)}")
        logger.error("Starting server anyway - will attempt to load model on first request")
    
    # Run the Flask app
    logger.info("Starting Flask server")
    app.run(host='0.0.0.0', port=5000, debug=False)