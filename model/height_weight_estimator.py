import tensorflow as tf
from tensorflow.keras.applications import MobileNetV2
from tensorflow.keras.models import Model
from tensorflow.keras.layers import Dense, GlobalAveragePooling2D, Dropout, Concatenate, Input

def create_weight_estimator(input_shape=(224, 224, 3)):
    """
    Creates a model for estimating weight from front image, side image, and height.
    Uses MobileNetV2 as the base model for image processing.
    
    Args:
        input_shape: Input shape of the images (default: 224x224x3)
        
    Returns:
        A compiled Keras model
    """
    # Create input layers
    front_image_input = Input(shape=input_shape, name='front_image')
    side_image_input = Input(shape=input_shape, name='side_image')
    height_input = Input(shape=(1,), name='height')
    
    # Load MobileNetV2 as base model (without top layers) for image processing
    base_model = MobileNetV2(input_shape=input_shape, include_top=False, weights='imagenet')
    
    # Freeze the base model layers
    base_model.trainable = False
    
    # Process front image
    front_features = base_model(front_image_input)
    front_features = GlobalAveragePooling2D()(front_features)
    front_features = Dense(512, activation='relu')(front_features)
    front_features = Dropout(0.3)(front_features)
    
    # Process side image
    side_features = base_model(side_image_input)
    side_features = GlobalAveragePooling2D()(side_features)
    side_features = Dense(512, activation='relu')(side_features)
    side_features = Dropout(0.3)(side_features)
    
    # Combine image features
    combined_features = Concatenate()([front_features, side_features])
    combined_features = Dense(256, activation='relu')(combined_features)
    combined_features = Dropout(0.3)(combined_features)
    
    # Add height input to the combined features
    height_features = Dense(64, activation='relu')(height_input)
    all_features = Concatenate()([combined_features, height_features])
    
    # Add final dense layers
    all_features = Dense(128, activation='relu')(all_features)
    all_features = Dropout(0.2)(all_features)
    
    # Output layer for weight prediction
    weight_output = Dense(1, name='weight')(all_features)
    
    # Create the model
    model = Model(inputs=[front_image_input, side_image_input, height_input], outputs=weight_output)
    
    # Compile the model
    model.compile(
        optimizer='adam',
        loss='mean_squared_error',
        metrics=['mean_absolute_error']
    )
    
    return model

def fine_tune_model(model, learning_rate=1e-4, num_layers_to_unfreeze=50):
    """
    Fine-tune the model by unfreezing some of the top layers of the base model.
    
    Args:
        model: The compiled model
        learning_rate: Learning rate for fine-tuning
        num_layers_to_unfreeze: Number of top layers to unfreeze in the base model
    
    Returns:
        The fine-tuned model
    """
    # Get the base model (MobileNetV2) which is used twice in our model
    # We need to find the layers that correspond to the base model
    # The base model is applied to both front and side images
    
    # Collect all MobileNetV2 layers that are used in the model
    mobilenet_layers = []
    for layer in model.layers:
        if isinstance(layer, tf.keras.Model) and "mobilenetv2" in layer.name.lower():
            mobilenet_layers.append(layer)
    
    # Unfreeze the top layers of each MobileNetV2 instance
    for base_model in mobilenet_layers:
        for layer in base_model.layers[-num_layers_to_unfreeze:]:
            layer.trainable = True
    
    # Recompile the model with a lower learning rate
    model.compile(
        optimizer=tf.keras.optimizers.Adam(learning_rate=learning_rate),
        loss='mean_squared_error',
        metrics=['mean_absolute_error']
    )
    
    return model

def preprocess_image(image):
    """
    Preprocess an image for model input.
    
    Args:
        image: Input image
        
    Returns:
        Preprocessed image
    """
    # Resize the image to the required input shape
    image = tf.image.resize(image, (224, 224))
    
    # Preprocess using MobileNetV2's preprocessing function
    image = tf.keras.applications.mobilenet_v2.preprocess_input(image)
    
    return image