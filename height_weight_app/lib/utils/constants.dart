import 'package:flutter/foundation.dart' show kIsWeb;
import 'dart:io' show Platform;

class Constants {
  // API endpoints
  static String get apiBaseUrl {
    if (kIsWeb) {
      // For web browsers
      return 'http://localhost:8080/api';
    } else if (!kIsWeb && Platform.isAndroid) {
      // For Android emulator
      return 'http://10.0.2.2:8080/api';
    } else if (!kIsWeb && Platform.isIOS) {
      // For iOS simulator
      return 'http://localhost:8080/api';
    } else {
      // Default fallback
      return 'http://localhost:8080/api';
    }
  }

  // Image capture settings
  static const double maxImageWidth = 800.0;
  static const double maxImageHeight = 1200.0;
  static const int imageQuality = 85;
  
  // UI constants
  static const double defaultPadding = 16.0;
  static const double smallPadding = 8.0;
  static const double mediumPadding = 16.0;
  static const double largePadding = 24.0;
}