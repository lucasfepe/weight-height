# Height and Weight Estimator from Images

This project implements a TensorFlow model for estimating a person's height and weight from images using MobileNetV2 as the base model.

## Project Structure

- `height_weight_estimator.py`: Contains the model architecture definition
- `train.py`: Script for training the model
- `predict.py`: Script for making predictions using a trained model
- `requirements.txt`: List of dependencies

## Setup

1. Install the required dependencies:

```bash
pip install -r requirements.txt
```

## Training

To train the model, you need a dataset of images with corresponding height and weight labels. The default training script contains a placeholder data loader that you need to replace with your actual data loading logic.

```bash
python train.py --data-dir /path/to/your/dataset --output-dir ./saved_models --batch-size 32 --epochs 50
```

The training process has two phases:
1. Initial training: Only the top layers are trained while the MobileNetV2 base is frozen
2. Fine-tuning: Some layers of the base model are unfrozen for fine-tuning

## Making Predictions

Once you have a trained model, you can use it to predict height and weight from new images:

```bash
python predict.py --model ./saved_models/height_weight_estimator_final.h5 --image /path/to/your/image.jpg
```

## Model Architecture

The model uses MobileNetV2 pretrained on ImageNet as the base feature extractor, followed by:
- Global Average Pooling
- Dense layers with dropout for regularization
- Two output heads: one for height and one for weight prediction

## Notes for Data Preparation

For best results:
- Images should be full-body shots
- Consistent camera distance and angle
- Include a diverse set of body types, heights, and weights
- Normalize height and weight values during training

## Limitations

This model is for demonstration purposes and may require:
- Extensive data collection
- Calibration based on camera position/angle
- Additional features like gender, age, etc. for improved accuracy 