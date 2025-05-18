import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:http/http.dart' as http;
import 'package:flutter/foundation.dart';
import '../utils/constants.dart';
import 'package:http_parser/http_parser.dart';

class ApiResponse {
  final bool success;
  final dynamic data;
  final String? message;

  ApiResponse({
    required this.success,
    this.data,
    this.message,
  });

  factory ApiResponse.fromJson(Map<String, dynamic> json) {
    return ApiResponse(
      success: json['success'] ?? false,
      data: json['data'],
      message: json['message'],
    );
  }
}

class WeightResult {
  final double weight;

  WeightResult({
    required this.weight,
  });

  factory WeightResult.fromJson(Map<String, dynamic> json) {
    return WeightResult(
      weight: json['weight']?.toDouble() ?? 0.0,
    );
  }
}

class ApiService {
  final http.Client _client = http.Client();
  
  // Upload method for non-web platforms
  Future<ApiResponse> uploadImage(File imageFile) async {
    try {
      // Create multipart request
      final request = http.MultipartRequest('POST', Uri.parse('${Constants.apiBaseUrl}/upload'));
      
      // Add the image file to the request
      final fileStream = http.ByteStream(imageFile.openRead());
      final fileLength = await imageFile.length();
      
      final multipartFile = http.MultipartFile(
        'image',
        fileStream,
        fileLength,
        filename: 'image.jpg',
      );
      
      request.files.add(multipartFile);
      
      // Send the request
      final streamedResponse = await request.send();
      final response = await http.Response.fromStream(streamedResponse);
      
      // Check if successful
      if (response.statusCode == 200) {
        final jsonResponse = json.decode(response.body);
        return ApiResponse.fromJson(jsonResponse);
      } else {
        return ApiResponse(
          success: false,
          message: 'Failed to upload image. Status code: ${response.statusCode}',
        );
      }
    } catch (e) {
      debugPrint('Error uploading image: $e');
      return ApiResponse(
        success: false,
        message: 'Error uploading image: $e',
      );
    }
  }

  // Upload method for web platforms
  Future<ApiResponse> uploadImageBytes(Uint8List imageBytes) async {
    try {
      // Create multipart request
      final request = http.MultipartRequest('POST', Uri.parse('${Constants.apiBaseUrl}/upload'));
      
      // Add the image bytes to the request
      final multipartFile = http.MultipartFile.fromBytes(
        'image',
        imageBytes,
        filename: 'image.jpg',
        contentType: MediaType('image', 'jpeg'),
      );
      
      request.files.add(multipartFile);
      
      // Send the request
      final streamedResponse = await request.send();
      final response = await http.Response.fromStream(streamedResponse);
      
      // Check if successful
      if (response.statusCode == 200) {
        final jsonResponse = json.decode(response.body);
        return ApiResponse.fromJson(jsonResponse);
      } else {
        return ApiResponse(
          success: false,
          message: 'Failed to upload image. Status code: ${response.statusCode}',
        );
      }
    } catch (e) {
      debugPrint('Error uploading image: $e');
      return ApiResponse(
        success: false,
        message: 'Error uploading image: $e',
      );
    }
  }

  // Get estimation results for a specific image ID
  Future<ApiResponse> getEstimationResults(String imageId) async {
    try {
      final response = await _client.get(
        Uri.parse('${Constants.apiBaseUrl}/estimate/$imageId'),
      );
      
      if (response.statusCode == 200) {
        final jsonResponse = json.decode(response.body);
        return ApiResponse.fromJson(jsonResponse);
      } else {
        return ApiResponse(
          success: false,
          message: 'Failed to get estimation results. Status code: ${response.statusCode}',
        );
      }
    } catch (e) {
      debugPrint('Error getting estimation results: $e');
      return ApiResponse(
        success: false,
        message: 'Error getting estimation results: $e',
      );
    }
  }

  // Upload images method for mobile platforms (uses File)
  Future<ApiResponse> uploadImages(File frontImage, File sideImage, double height) async {
    try {
      // Create multipart request
      final request = http.MultipartRequest('POST', Uri.parse('${Constants.apiBaseUrl}/estimate-weight'));
      
      // Add the front image file
      final frontFileStream = http.ByteStream(frontImage.openRead());
      final frontFileLength = await frontImage.length();
      
      final frontMultipartFile = http.MultipartFile(
        'front_image',
        frontFileStream,
        frontFileLength,
        filename: 'front_image.jpg',
      );
      
      // Add the side image file
      final sideFileStream = http.ByteStream(sideImage.openRead());
      final sideFileLength = await sideImage.length();
      
      final sideMultipartFile = http.MultipartFile(
        'side_image',
        sideFileStream,
        sideFileLength,
        filename: 'side_image.jpg',
      );
      
      // Add height value
      request.fields['height'] = height.toString();
      
      // Add both files to request
      request.files.add(frontMultipartFile);
      request.files.add(sideMultipartFile);
      
      // Send the request
      final streamedResponse = await request.send();
      final response = await http.Response.fromStream(streamedResponse);
      
      // Check if successful
      if (response.statusCode == 200) {
        final jsonResponse = json.decode(response.body);
        return ApiResponse.fromJson(jsonResponse);
      } else {
        return ApiResponse(
          success: false,
          message: 'Failed to process request. Status code: ${response.statusCode}',
        );
      }
    } catch (e) {
      debugPrint('Error processing request: $e');
      return ApiResponse(
        success: false,
        message: 'Error processing request: $e',
      );
    }
  }

  // Upload images method for web platforms (uses Uint8List)
  Future<ApiResponse> uploadImagesWeb(Uint8List frontImageBytes, Uint8List sideImageBytes, double height) async {
    try {
      // Create multipart request
      final request = http.MultipartRequest('POST', Uri.parse('${Constants.apiBaseUrl}/estimate-weight'));
      
      // Add the front image bytes
      final frontMultipartFile = http.MultipartFile.fromBytes(
        'front_image',
        frontImageBytes,
        filename: 'front_image.jpg',
        contentType: MediaType('image', 'jpeg'),
      );
      
      // Add the side image bytes
      final sideMultipartFile = http.MultipartFile.fromBytes(
        'side_image',
        sideImageBytes,
        filename: 'side_image.jpg',
        contentType: MediaType('image', 'jpeg'),
      );
      
      // Add height value
      request.fields['height'] = height.toString();
      
      // Add both files to request
      request.files.add(frontMultipartFile);
      request.files.add(sideMultipartFile);
      
      // Send the request
      final streamedResponse = await request.send();
      final response = await http.Response.fromStream(streamedResponse);
      
      // Check if successful
      if (response.statusCode == 200) {
        final jsonResponse = json.decode(response.body);
        return ApiResponse.fromJson(jsonResponse);
      } else {
        return ApiResponse(
          success: false,
          message: 'Failed to process request. Status code: ${response.statusCode}',
        );
      }
    } catch (e) {
      debugPrint('Error processing request: $e');
      return ApiResponse(
        success: false,
        message: 'Error processing request: $e',
      );
    }
  }

  // Weight prediction for mobile platforms
  Future<WeightResult> predictWeight(File frontImage, File sideImage, double height) async {
    final response = await uploadImages(frontImage, sideImage, height);
    
    if (response.success && response.data != null) {
      return WeightResult.fromJson(response.data);
    }
    
    throw Exception(response.message ?? 'Failed to predict weight');
  }

  // Weight prediction for web platforms
  // Add these in predictWeightWeb method:
Future<WeightResult> predictWeightWeb(Uint8List frontImageBytes, Uint8List sideImageBytes, double height) async {
  print("🌐 predictWeightWeb called with:");
  print("  - frontImageBytes: ${frontImageBytes.length} bytes");
  print("  - sideImageBytes: ${sideImageBytes.length} bytes");
  print("  - height: $height cm");
  
  try {
    final response = await uploadImagesWeb(frontImageBytes, sideImageBytes, height);
    
    print("✅ API response received: success=${response.success}");
    
    if (response.success && response.data != null) {
      final result = WeightResult.fromJson(response.data);
      print("✅ Weight result: ${result.weight}kg");
      return result;
    }
    
    print("❌ API error: ${response.message}");
    throw Exception(response.message ?? 'Failed to predict weight');
  } catch (e) {
    print("❌ Exception in predictWeightWeb: $e");
    throw e;
  }
}
}