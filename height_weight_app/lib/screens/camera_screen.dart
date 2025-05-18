import 'dart:io';
import 'package:flutter/material.dart';
import 'package:camera/camera.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as path;
import 'package:permission_handler/permission_handler.dart';

class CameraScreen extends StatefulWidget {
  final String imageType; // 'front' or 'side'
  
  const CameraScreen({
    super.key,
    required this.imageType,
  });

  @override
  State<CameraScreen> createState() => _CameraScreenState();
}

class _CameraScreenState extends State<CameraScreen> with WidgetsBindingObserver {
  CameraController? _controller;
  List<CameraDescription> _cameras = [];
  bool _isCameraInitialized = false;
  bool _isPermissionGranted = false;
  
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _requestPermissions();
  }
  
  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _controller?.dispose();
    super.dispose();
  }
  
  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (_controller == null || !_controller!.value.isInitialized) {
      return;
    }
    
    if (state == AppLifecycleState.inactive) {
      _controller?.dispose();
    } else if (state == AppLifecycleState.resumed) {
      _initializeCamera();
    }
  }
  
  Future<void> _requestPermissions() async {
    final status = await Permission.camera.request();
    setState(() {
      _isPermissionGranted = status.isGranted;
    });
    
    if (_isPermissionGranted) {
      await _initializeCamera();
    }
  }
  
  Future<void> _initializeCamera() async {
    _cameras = await availableCameras();
    
    if (_cameras.isEmpty) {
      return;
    }
    
    // Use the first camera from the list (usually the back camera)
    final CameraDescription camera = _cameras.first;
    
    _controller = CameraController(
      camera,
      ResolutionPreset.high,
      enableAudio: false,
    );
    
    try {
      await _controller!.initialize();
      setState(() {
        _isCameraInitialized = true;
      });
    } catch (e) {
      debugPrint('Error initializing camera: $e');
    }
  }
  
  Future<void> _takePicture() async {
    if (_controller == null || !_controller!.value.isInitialized) {
      return;
    }
    
    try {
      // Take picture and get image file
      final XFile image = await _controller!.takePicture();
      
      // Create a more permanent file
      final Directory appDir = await getApplicationDocumentsDirectory();
      final String fileName = path.basename(image.path);
      final String filePath = path.join(appDir.path, fileName);
      
      // Copy the temporary file to app directory
      final File savedImage = await File(image.path).copy(filePath);
      
      // Return the file to the previous screen
      if (!mounted) return;
      Navigator.of(context).pop(savedImage);
    } catch (e) {
      debugPrint('Error taking picture: $e');
    }
  }
  
  @override
  Widget build(BuildContext context) {
    if (!_isPermissionGranted) {
      return Scaffold(
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Text('Camera permission is required'),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: _requestPermissions,
                child: const Text('Request Permission'),
              ),
            ],
          ),
        ),
      );
    }
    
    if (!_isCameraInitialized || _controller == null) {
      return const Scaffold(
        body: Center(
          child: CircularProgressIndicator(),
        ),
      );
    }
    
    return Scaffold(
      appBar: AppBar(
        title: Text('Take ${widget.imageType.capitalize()} Photo'),
        centerTitle: true,
      ),
      body: Stack(
        children: [
          // Camera preview
          CameraPreview(_controller!),
          
          // Capture button
          Positioned(
            bottom: 30,
            left: 0,
            right: 0,
            child: Center(
              child: GestureDetector(
                onTap: _takePicture,
                child: Container(
                  height: 80,
                  width: 80,
                  decoration: BoxDecoration(
                    color: Colors.white.withOpacity(0.8),
                    shape: BoxShape.circle,
                  ),
                  child: const Center(
                    child: Icon(
                      Icons.camera,
                      size: 40,
                      color: Colors.black,
                    ),
                  ),
                ),
              ),
            ),
          ),
          
          // Close button
          Positioned(
            top: 40,
            right: 20,
            child: CircleAvatar(
              backgroundColor: Colors.black.withOpacity(0.5),
              child: IconButton(
                icon: const Icon(Icons.close, color: Colors.white),
                onPressed: () => Navigator.of(context).pop(),
              ),
            ),
          ),
          
          // Guidelines overlay
          Positioned.fill(
            child: CustomPaint(
              painter: GuidePainter(isFrontView: widget.imageType == 'front'),
            ),
          ),
          
          // Instructions
          Positioned(
            bottom: 120,
            left: 0,
            right: 0,
            child: Center(
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.6),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  widget.imageType == 'front' 
                      ? 'Face the camera directly'
                      : 'Stand sideways to the camera',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 16,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

// Extension to capitalize first letter
extension StringExtension on String {
  String capitalize() {
    return "${this[0].toUpperCase()}${substring(1)}";
  }
}

// Custom painter for drawing guidelines
class GuidePainter extends CustomPainter {
  final bool isFrontView;
  
  GuidePainter({required this.isFrontView});
  
  @override
  void paint(Canvas canvas, Size size) {
    final Paint paint = Paint()
      ..color = Colors.white.withOpacity(0.5)
      ..strokeWidth = 1.5
      ..style = PaintingStyle.stroke;
    
    if (isFrontView) {
      // Front view guidelines
      final double centerX = size.width / 2;
      // Vertical line down the middle
      canvas.drawLine(
        Offset(centerX, size.height * 0.2),
        Offset(centerX, size.height * 0.8),
        paint,
      );
      
      // Draw horizontal guides
      final double headY = size.height * 0.2;
      final double shoulderY = size.height * 0.3;
      final double waistY = size.height * 0.5;
      final double kneeY = size.height * 0.7;
      
      // Horizontal lines
      canvas.drawLine(
        Offset(centerX - 30, headY),
        Offset(centerX + 30, headY),
        paint,
      );
      
      canvas.drawLine(
        Offset(centerX - 50, shoulderY),
        Offset(centerX + 50, shoulderY),
        paint,
      );
      
      canvas.drawLine(
        Offset(centerX - 40, waistY),
        Offset(centerX + 40, waistY),
        paint,
      );
      
      canvas.drawLine(
        Offset(centerX - 30, kneeY),
        Offset(centerX + 30, kneeY),
        paint,
      );
    } else {
      // Side view guidelines - vertical outline
      final double centerX = size.width / 2;
      
      // Vertical body outline
      final Path outlinePath = Path()
        ..moveTo(centerX - 20, size.height * 0.2)  // Head top
        ..lineTo(centerX + 20, size.height * 0.2)  // Head top
        ..lineTo(centerX + 20, size.height * 0.3)  // Neck
        ..lineTo(centerX + 40, size.height * 0.5)  // Body
        ..lineTo(centerX + 20, size.height * 0.8)  // Legs
        ..lineTo(centerX - 20, size.height * 0.8)  // Feet
        ..lineTo(centerX - 10, size.height * 0.5)  // Back
        ..lineTo(centerX - 20, size.height * 0.3)  // Back neck
        ..close();
        
      canvas.drawPath(outlinePath, paint);
    }
  }
  
  @override
  bool shouldRepaint(CustomPainter oldDelegate) => false;
} 