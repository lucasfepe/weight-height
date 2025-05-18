import tensorflow as tf
import numpy as np
import os
import requests
import json
from tensorflow.keras.callbacks import ModelCheckpoint, EarlyStopping, ReduceLROnPlateau
from height_weight_estimator import create_weight_estimator, fine_tune_model, preprocess_image

def load_training_data_from_api(api_url):
    """
    Load training data from the API endpoint.
    
    Args:
        api_url: Base URL of the API
        
    Returns:
        List of training data records
    """
    export_url = f"{api_url}/api/export-training-data"
    
    try:
        response = requests.get(export_url)
        response.raise_for_status()  # Raise an exception for HTTP errors
        
        data = response.json()
        
        if data.get("success", False) and "data" in data:
            print(f"Successfully loaded {len(data['data'])} training records from API")
            return data["data"]
        else:
            print(f"API returned error: {data.get('message', 'Unknown error')}")
            return []
    except Exception as e:
        print(f"Failed to load training data from API: {e}")
        return []

def load_and_prepare_data(data_dir, batch_size=32, split_ratio=0.2, api_url=None):
    """
    Load and prepare the dataset for training.
    
    Args:
        data_dir: Directory containing the dataset
        batch_size: Batch size for training
        split_ratio: Validation split ratio
        api_url: Base URL of the API to fetch training data
        
    Returns:
        Training and validation datasets
    """
    # If API URL is provided, try to load data from API
    training_records = []
    if api_url:
        training_records = load_training_data_from_api(api_url)

    # If we got data from the API
    if training_records:
        # Create dataset from API data
        front_images = []
        side_images = []
        heights = []
        weights = []
        
        for record in training_records:
            # Load and preprocess images
            try:
                front_img_path = record["front_image_path"]
                side_img_path = record["side_image_path"]
                
                # If paths are relative to API, make them absolute using data_dir
                if not os.path.isabs(front_img_path):
                    front_img_path = os.path.join(data_dir, front_img_path)
                if not os.path.isabs(side_img_path):
                    side_img_path = os.path.join(data_dir, side_img_path)
                
                # Process images if they exist
                if os.path.exists(front_img_path) and os.path.exists(side_img_path):
                    front_img = preprocess_image(front_img_path)
                    side_img = preprocess_image(side_img_path)
                    
                    front_images.append(front_img)
                    side_images.append(side_img)
                    heights.append([record["height"]])
                    weights.append([record["actual_weight"]])
            except Exception as e:
                print(f"Error processing record: {e}")
                continue
        
        # Convert lists to numpy arrays
        if front_images:
            front_images = np.array(front_images)
            side_images = np.array(side_images)
            heights = np.array(heights)
            weights = np.array(weights)
            
            # Split data into training and validation sets
            indices = np.random.permutation(len(weights))
            split_idx = int(len(indices) * (1 - split_ratio))
            
            train_indices = indices[:split_idx]
            val_indices = indices[split_idx:]
            
            # Create TensorFlow datasets
            train_dataset = tf.data.Dataset.from_tensor_slices((
                {
                    'front_image': front_images[train_indices],
                    'side_image': side_images[train_indices],
                    'height': heights[train_indices]
                }, 
                weights[train_indices]
            )).batch(batch_size)
            
            val_dataset = tf.data.Dataset.from_tensor_slices((
                {
                    'front_image': front_images[val_indices],
                    'side_image': side_images[val_indices],
                    'height': heights[val_indices]
                }, 
                weights[val_indices]
            )).batch(batch_size)
            
            print(f"Created dataset with {len(train_indices)} training samples and {len(val_indices)} validation samples")
            return train_dataset, val_dataset, len(train_indices) // batch_size + 1
    
    print("No data loaded from API or insufficient records. Using dummy data instead.")
    
    # If no data from API or not enough records, fall back to dummy data
    def dummy_data_generator(batch_size):
        while True:
            # Generate random front and side images
            front_images = np.random.random((batch_size, 224, 224, 3))
            side_images = np.random.random((batch_size, 224, 224, 3))
            
            # Generate heights and weights
            heights = np.random.normal(170, 10, (batch_size, 1))
            weights = np.random.normal(70, 15, (batch_size, 1))
            
            # Return inputs and target (weight only)
            yield {
                'front_image': front_images,
                'side_image': side_images,
                'height': heights
            }, weights
    
    # Create dummy train and validation datasets
    steps_per_epoch = 100  # Arbitrary number for demonstration
    train_dataset = dummy_data_generator(batch_size)
    val_dataset = dummy_data_generator(batch_size)
    
    return train_dataset, val_dataset, steps_per_epoch

def train_model(data_dir, output_dir, api_url=None, batch_size=32, epochs=50):
    """
    Train the weight estimation model.
    
    Args:
        data_dir: Directory containing the dataset
        output_dir: Directory to save model checkpoints
        api_url: Base URL of the API to fetch training data
        batch_size: Batch size for training
        epochs: Number of epochs to train
    """
    # Create the model
    model = create_weight_estimator()
    print("Model created successfully")
    
    # Load and prepare data
    train_dataset, val_dataset, steps_per_epoch = load_and_prepare_data(
        data_dir, batch_size, api_url=api_url)
    
    # Set up callbacks
    checkpoint = ModelCheckpoint(
        f"{output_dir}/weight_model_checkpoint.h5",
        save_best_only=True,
        monitor='val_loss',
        mode='min'
    )
    
    early_stopping = EarlyStopping(
        monitor='val_loss',
        patience=10,
        restore_best_weights=True
    )
    
    reduce_lr = ReduceLROnPlateau(
        monitor='val_loss',
        factor=0.2,
        patience=5,
        min_lr=1e-6
    )
    
    callbacks = [checkpoint, early_stopping, reduce_lr]
    
    # Initial training phase - train only the top layers
    print("Starting initial training phase...")
    model.fit(
        train_dataset,
        steps_per_epoch=steps_per_epoch,
        validation_data=val_dataset,
        validation_steps=steps_per_epoch // 5,
        epochs=epochs // 2,
        callbacks=callbacks
    )
    
    # Fine-tuning phase - unfreeze some layers of the base model
    print("Starting fine-tuning phase...")
    model = fine_tune_model(model)
    
    model.fit(
        train_dataset,
        steps_per_epoch=steps_per_epoch,
        validation_data=val_dataset,
        validation_steps=steps_per_epoch // 5,
        epochs=epochs // 2,
        callbacks=callbacks
    )
    
    # Save the final model
    model.save(f"{output_dir}/weight_estimator_final.h5")
    print(f"Model training complete. Final model saved to {output_dir}/weight_estimator_final.h5")

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description='Train weight estimation model')
    parser.add_argument('--data-dir', type=str, required=True, help='Directory containing the dataset')
    parser.add_argument('--output-dir', type=str, default='./saved_models', help='Directory to save model checkpoints')
    parser.add_argument('--api-url', type=str, default=None, help='Base URL of the API to fetch training data')
    parser.add_argument('--batch-size', type=int, default=32, help='Batch size for training')
    parser.add_argument('--epochs', type=int, default=50, help='Number of epochs to train')
    
    args = parser.parse_args()
    
    # Create output directory if it doesn't exist
    import os
    os.makedirs(args.output_dir, exist_ok=True)
    
    train_model(args.data_dir, args.output_dir, args.api_url, args.batch_size, args.epochs) 