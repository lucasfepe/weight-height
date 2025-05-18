import 'dart:io' as io;
import 'dart:typed_data';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';
import '../services/api_service.dart';
import '../utils/constants.dart';
import '../widgets/result_display.dart';
import 'camera_screen.dart';
import 'training_data_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  // For non-web platforms
  io.File? _frontImageFile;
  io.File? _sideImageFile;
  
  // For web platforms
  Uint8List? _frontImageBytes;
  Uint8List? _sideImageBytes;
  
  bool _isLoading = false;
  WeightResult? _result;
  final TextEditingController _heightController = TextEditingController();
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
        _result = null;
      });
      _printDebugState("After camera capture");
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
              _result = null;
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
            _result = null;
          });
        }
        _printDebugState("After picking image");
      }
    } catch (e) {
      print("⚠️ Error picking image: $e");
    }
  }

  void _printDebugState(String context) {
    print("🔍 DEBUG STATE ($context):");
    print("  - Platform: ${kIsWeb ? 'Web' : 'Mobile/Desktop'}");
    print("  - hasFrontImage: $hasFrontImage");
    print("  - hasSideImage: $hasSideImage");
    print("  - Height entered: ${_heightController.text.isEmpty ? 'Empty' : _heightController.text}");
    print("  - Is loading: $_isLoading");
    print("  - Button should be: ${(!hasFrontImage || !hasSideImage || _heightController.text.isEmpty || _isLoading) ? 'DISABLED' : 'ENABLED'}");
    
    if (kIsWeb) {
      print("  - _frontImageBytes: ${_frontImageBytes != null ? '${_frontImageBytes!.length} bytes' : 'null'}");
      print("  - _sideImageBytes: ${_sideImageBytes != null ? '${_sideImageBytes!.length} bytes' : 'null'}");
    } else {
      print("  - _frontImageFile: $_frontImageFile");
      print("  - _sideImageFile: $_sideImageFile");
    }
  }

  Future<void> _analyzeImages() async {
    print("🔍 _analyzeImages called");
    
    if (!hasFrontImage || !hasSideImage) {
      print("⚠️ Missing images: front=$hasFrontImage, side=$hasSideImage");
      return;
    }
    
    // Validate height input
    if (_formKey.currentState == null || !_formKey.currentState!.validate()) {
      print("⚠️ Height validation failed");
      return;
    }

    // Parse height value
    final double height = double.parse(_heightController.text);
    print("📏 Height parsed: $height cm");
    
    setState(() {
      _isLoading = true;
      print("⏳ Loading state set to true");
    });

    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      
      print("🌐 Calling API for weight prediction on ${kIsWeb ? 'web' : 'mobile'}");
      WeightResult result;
      if (kIsWeb) {
        // Web implementation using bytes
        print("🌐 Using web implementation with bytes");
        result = await apiService.predictWeightWeb(_frontImageBytes!, _sideImageBytes!, height);
      } else {
        // Mobile implementation using File
        print("📱 Using mobile implementation with File");
        result = await apiService.predictWeight(_frontImageFile!, _sideImageFile!, height);
      }

      print("✅ Got result: weight=${result.weight}kg");
      setState(() {
        _result = result;
        _isLoading = false;
        print("⏳ Loading state set to false");
      });
    } catch (e) {
      print("❌ Error during analysis: $e");
      setState(() {
        _isLoading = false;
        print("⏳ Loading state set to false after error");
      });
      
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    // Debug print each time build is called
    _printDebugState("In build method");
    
    return Scaffold(
      appBar: AppBar(
        title: const Text('Weight Estimator'),
        centerTitle: true,
        actions: [
          IconButton(
            icon: const Icon(Icons.science),
            tooltip: 'Contribute Training Data',
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (context) => const TrainingDataScreen(),
                ),
              );
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(Constants.defaultPadding),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const SizedBox(height: Constants.smallPadding),
              const Text(
                'Please provide a front and side photo, along with your height to estimate weight',
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
              
              // Height input field
              TextFormField(
                controller: _heightController,
                decoration: const InputDecoration(
                  labelText: 'Height (cm)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.height),
                ),
                keyboardType: TextInputType.number,
                inputFormatters: [
                  FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d*')),
                ],
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Please enter your height';
                  }
                  final double? height = double.tryParse(value);
                  if (height == null || height <= 0 || height > 250) {
                    return 'Please enter a valid height';
                  }
                  return null;
                },
                onChanged: (value) {
                  print("📏 Height changed to: $value");
                  // Force a rebuild to update button state
                  setState(() {});
                },
              ),
              
              const SizedBox(height: Constants.largePadding),
              
              // Analyze button
              ElevatedButton(
                onPressed: (!hasFrontImage || !hasSideImage || _heightController.text.isEmpty || _isLoading) 
                    ? null 
                    : _analyzeImages,
                style: ElevatedButton.styleFrom(
                  backgroundColor: Theme.of(context).primaryColor,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                ),
                child: _isLoading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          color: Colors.white,
                          strokeWidth: 2,
                        ),
                      )
                    : const Text('ANALYZE'),
              ),
              
              // Debug button - explicitly show current state
              ElevatedButton(
                onPressed: () {
                  _printDebugState("Debug button pressed");
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text(
                      'DEBUG: Front: $hasFrontImage, Side: $hasSideImage, Height: ${_heightController.text.isNotEmpty}'
                    )),
                  );
                },
                child: const Text('DEBUG STATE'),
              ),
              
              if (_result != null)
                Padding(
                  padding: const EdgeInsets.only(top: Constants.largePadding),
                  child: ResultDisplay(result: _result!),
                ),
            ],
          ),
        ),
      ),
    );
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
          style: const TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: Constants.smallPadding),
        _buildImagePreview(image),
        const SizedBox(height: Constants.smallPadding),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            ElevatedButton.icon(
              onPressed: kIsWeb ? null : onCameraPressed,
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

  Widget _buildImagePreview(dynamic image) {
    if (image == null) {
      return Container(
        height: 200,
        decoration: BoxDecoration(
          color: Colors.grey[200],
          borderRadius: BorderRadius.circular(12),
        ),
        child: const Center(
          child: Icon(
            Icons.person,
            size: 80,
            color: Colors.grey,
          ),
        ),
      );
    }

    return ClipRRect(
      borderRadius: BorderRadius.circular(12),
      child: kIsWeb
          ? Image.memory(
              image as Uint8List,
              height: 200,
              width: double.infinity,
              fit: BoxFit.cover,
            )
          : Image.file(
              image as io.File,
              height: 200,
              width: double.infinity,
              fit: BoxFit.cover,
            ),
    );
  }
}