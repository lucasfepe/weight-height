from height_weight_estimator import create_height_weight_estimator
import os
import tensorflow as tf

def create_and_save_dummy_model():
    # Create the model using your existing architecture
    model = create_height_weight_estimator()
    
    # Create directory to save the model
    model_dir = 'models'
    os.makedirs(model_dir, exist_ok=True)
    
    # Save the model using .h5 format
    model_path = os.path.join(model_dir, 'height_weight_model.h5')
    model.save(model_path)
    
    print(f"Dummy dual-input model created and saved to: {model_path}")
    print(f"You can now run: set MODEL_PATH={model_path}")
    print("python ml_service.py")
    
    return model_path

if __name__ == "__main__":
    create_and_save_dummy_model()