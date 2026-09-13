import 'package:flutter/material.dart';
import 'package:grpc/grpc.dart';
import 'package:logger/logger.dart';

import 'data/service.dart';
import 'data/token.dart';
import 'generated/generated.dart';

/// Holds the authentication state for the web UI.
class AuthService extends ChangeNotifier {
  final Backend backend;
  final TokenStore tokenStore;
  final Logger _logger = Logger();

  bool _isAuthenticated = false;

  bool get isAuthenticated => _isAuthenticated;

  AuthService({required this.backend, required this.tokenStore}) {
    tokenStore.onUnauthorized = _handleUnauthorized;
  }

  /// Logs in and returns an error message, or null on success.
  Future<String?> login(String username, String password) async {
    try {
      final response = await backend.auth.login(
        LoginRequest(username: username, password: password),
      );

      if (!response.success || response.token.isEmpty) {
        return response.message.isEmpty ? 'Login failed' : response.message;
      }

      tokenStore.token = response.token;
      _isAuthenticated = true;
      _logger.i('User $username logged in');
      notifyListeners();
      return null;
    } on GrpcError catch (e) {
      _logger.e('Login failed: ${e.message}');
      return e.message ?? 'Login failed';
    } catch (e) {
      _logger.e('Login failed: $e');
      return 'Login failed';
    }
  }

  /// Registers a user and returns an error message, or null on success.
  Future<String?> register(
    String username,
    String password,
    String registerSecret,
  ) async {
    try {
      final response = await backend.auth.registerUser(
        RegisterRequest(
          username: username,
          password: password,
          registerSecret: registerSecret,
        ),
      );

      if (!response.success || response.token.isEmpty) {
        return response.message.isEmpty ? 'Registration failed' : response.message;
      }

      tokenStore.token = response.token;
      _isAuthenticated = true;
      _logger.i('User $username registered');
      notifyListeners();
      return null;
    } on GrpcError catch (e) {
      _logger.e('Registration failed: ${e.message}');
      return e.message ?? 'Registration failed';
    } catch (e) {
      _logger.e('Registration failed: $e');
      return 'Registration failed';
    }
  }

  void logout() {
    tokenStore.token = '';
    _isAuthenticated = false;
    _logger.i('User logged out');
    notifyListeners();
  }

  void _handleUnauthorized() {
    if (_isAuthenticated) {
      logout();
    }
  }

  static AuthService of(BuildContext context) =>
      context.dependOnInheritedWidgetOfExactType<AuthScope>()!.notifier!;
}

class AuthScope extends InheritedNotifier<AuthService> {
  const AuthScope({required super.notifier, required super.child, super.key});
}
