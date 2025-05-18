import 'dart:io' as io;
import 'dart:typed_data';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';
import '../services/api_service.dart';
import '../utils/constants.dart';
import 'camera_screen.dart';

class TrainingDataScreen extends StatefulWidget {
  const TrainingDataScreen({super.key});

  @override
  State<TrainingDataScreen> createState() => _TrainingDataScreenState();
}

class _TrainingDataScreenState extends State<TrainingDataScreen> {
  // For non-web platforms
  io.File? _frontImageFile;
  io.File? _sideImageFile;
  
  // For web platforms
  Uint8List? _frontImageBytes;
  Uint8List? _sideImageBytes;
  
  bool _isLoading = false;
  bool _isSaved = false;
  final TextEditingController _heightController = TextEditingController();
  final TextEditingController _weightController = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  // Helper getters to check if images are available
  bool get hasFrontImage => kIsWeb ? _frontImageBytes != null : _frontImageFile != null;
  bool get hasSideImage => kIsWeb ? _sideImageBytes != null : _sideImageFile != null;
  
  // Helper getters to retrieve the image in the appropriate format
  dynamic get frontImage => kIsWeb ? _frontImageBytes : _frontImageFile;
  dynamic get sideImage => kIsWeb ? _sideImageBytes : _sideImageFile;

  @override
  void dispose() {
    _heightController.dispose();
    _weightController.dispose();
    super.dispose();
  }

  Future<void> _takePicture(bool isFrontImage) async {
    if (kIsWeb) {
      // Web doesn't support direct camera access through this method
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Camera not supported on web. Please use Gallery instead.')),
      );
      return;
    }
    
    final String imageType = isFrontImage ? 'front' : 'side';
    
    final io.File? image = await Navigator.of(context).push<io.File?>(
      MaterialPageRoute(
        builder: (context) => CameraScreen(imageType: imageType),
      ),
    );

    if (image != null) {
      setState(() {
        if (isFrontImage) {
          _frontImageFile = image;
          print("📸 Front image set from camera: $_frontImageFile");
        } else {
          _sideImageFile = image;
          print("📸 Side image set from camera: $_sideImageFile");
        }
      });
    }
  }

  Future<void> _pickImage(bool isFrontImage) async {
    print("🖼️ Picking image for ${isFrontImage ? 'front' : 'side'} on ${kIsWeb ? 'web' : 'mobile'}");
    
    final ImagePicker picker = ImagePicker();
    try {
      final XFile? image = await picker.pickImage(
        source: ImageSource.gallery,
        maxWidth: Constants.maxImageWidth,
        maxHeight: Constants.maxImageHeight,
        imageQuality: Constants.imageQuality,
      );

      print("🖼️ Image picked: ${image != null ? 'success' : 'null or canceled'}");

      if (image != null) {
        if (kIsWeb) {
          // For web, read the image as bytes
          try {
            final bytes = await image.readAsBytes();
            print("🖼️ Web: Image bytes read, length: ${bytes.length} bytes");
            
            setState(() {
              if (isFrontImage) {
                _frontImageBytes = bytes;
                print("🖼️ Web: Front image bytes set, length: ${_frontImageBytes?.length}");
              } else {
                _sideImageBytes = bytes;
                print("🖼️ Web: Side image bytes set, length: ${_sideImageBytes?.length}");
              }
            });
          } catch (e) {
            print("⚠️ Error reading image bytes: $e");
          }
        } else {
          // For mobile/desktop
          setState(() {
            if (isFrontImage) {
              _frontImageFile = io.File(image.path);
              print("🖼️ Mobile: Front image file set: $_frontImageFile");
            } else {
              _sideImageFile = io.File(image.path);
              print("🖼️ Mobile: Side image file set: $_sideImageFile");
            }
          });
        }
      }
    } catch (e) {
      print("⚠️ Error picking image: $e");
    }
  }

  Future<void> _saveTrainingData() async {
    if (!hasFrontImage || !hasSideImage) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please provide both front and side images')),
      );
      return;
    }
    
    // Validate form
    if (_formKey.currentState == null || !_formKey.currentState!.validate()) {
      return;
    }

    // Parse values
    final double height = double.parse(_heightController.text);
    final double weight = double.parse(_weightController.text);
    
    setState(() {
      _isLoading = true;
    });

    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      
      ApiResponse response;
      if (kIsWeb) {
        // Web implementation using bytes
        response = await apiService.saveTrainingDataWeb(_frontImageBytes!, _sideImageBytes!, height, weight);
      } else {
        // Mobile implementation using File
        response = await apiService.saveTrainingData(_frontImageFile!, _sideImageFile!, height, weight);
      }

      if (response.success) {
        setState(() {
          _isSaved = true;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Training data saved successfully! Thank you for your contribution.')),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: ${response.message ?? "Unknown error"}')),
        );
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  void _resetForm() {
    setState(() {
      _frontImageFile = null;
      _sideImageFile = null;
      _frontImageBytes = null;
      _sideImageBytes = null;
      _heightController.clear();
      _weightController.clear();
      _isSaved = false;
    });
  }

  Widget _buildImageSection({
    required String title,
    required dynamic image,
    required VoidCallback onCameraPressed,
    required VoidCallback onGalleryPressed,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 8),
        Container(
          height: 200,
          width: double.infinity,
          decoration: BoxDecoration(
            border: Border.all(color: Colors.grey),
            borderRadius: BorderRadius.circular(8),
          ),
          child: (kIsWeb && image is Uint8List) 
            ? Image.memory(image, fit: BoxFit.cover)
            : (!kIsWeb && image is io.File) 
                ? Image.file(image, fit: BoxFit.cover)
                : Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.image, size: 48, color: Colors.grey),
                        const SizedBox(height: 8),
                        Text('No $title selected'),
                      ],
                    ),
                  ),
        ),
        const SizedBox(height: 8),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            ElevatedButton.icon(
              onPressed: onCameraPressed,
              icon: const Icon(Icons.camera_alt),
              label: const Text('Camera'),
            ),
            ElevatedButton.icon(
              onPressed: onGalleryPressed,
              icon: const Icon(Icons.photo_library),
              label: const Text('Gallery'),
            ),
          ],
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Add Training Data'),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(Constants.defaultPadding),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Contribute to Model Training',
                style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              const Text(
                'Help improve our model by providing your photos with actual height and weight',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 16),
              ),
              const SizedBox(height: Constants.largePadding),
              
              // Front image section
              _buildImageSection(
                title: 'Front Image',
                image: frontImage, 
                onCameraPressed: () => _takePicture(true),
                onGalleryPressed: () => _pickImage(true),
              ),
              
              const SizedBox(height: Constants.largePadding),
              
              // Side image section
              _buildImageSection(
                title: 'Side Image',
                image: sideImage, 
                onCameraPressed: () => _takePicture(false),
                onGalleryPressed: () => _pickImage(false),
              ),
              
              const SizedBox(height: Constants.largePadding),
              
              // Height input
              TextFormField(
                controller: _heightController,
                decoration: const InputDecoration(
                  labelText: 'Height (cm)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.height),
                ),
                keyboardType: TextInputType.number,
                inputFormatters: [
                  FilteringTextInputFormatter.allow(RegExp(r'(^\d*\.?\d*)')),
                ],
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Please enter your height';
                  }
                  final double? height = double.tryParse(value);
                  if (height == null || height <= 0) {
                    return 'Please enter a valid height';
                  }
                  return null;
                },
              ),
              
              const SizedBox(height: Constants.mediumPadding),
              
              // Weight input
              TextFormField(
                controller: _weightController,
                decoration: const InputDecoration(
                  labelText: 'Actual Weight (kg)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.monitor_weight),
                ),
                keyboardType: TextInputType.number,
                inputFormatters: [
                  FilteringTextInputFormatter.allow(RegExp(r'(^\d*\.?\d*)')),
                ],
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Please enter your actual weight';
                  }
                  final double? weight = double.tryParse(value);
                  if (weight == null || weight <= 0) {
                    return 'Please enter a valid weight';
                  }
                  return null;
                },
              ),
              
              const SizedBox(height: Constants.largePadding),
              
              // Save button
              ElevatedButton.icon(
                onPressed: _isLoading || _isSaved ? null : _saveTrainingData,
                icon: _isLoading 
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : _isSaved 
                      ? const Icon(Icons.check)
                      : const Icon(Icons.save),
                label: Text(_isLoading 
                  ? 'Saving...' 
                  : _isSaved 
                      ? 'Saved Successfully'
                      : 'Save Training Data'
                ),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
              ),
              
              if (_isSaved) ...[
                const SizedBox(height: Constants.mediumPadding),
                OutlinedButton.icon(
                  onPressed: _resetForm,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Add Another'),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 16),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
} 