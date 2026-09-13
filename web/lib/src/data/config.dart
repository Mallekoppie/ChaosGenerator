import 'package:web/web.dart' as web;

/// Resolves the address that the gRPC-Web endpoints are served from.
///
/// When the app is served by the master itself (for example through a
/// `kubectl port-forward`) the origin of the current page is used. During local
/// development the value can be overridden, for example:
///
///   flutter run -d chrome --dart-define=CHAOS_MASTER_URL=http://localhost:8080
class AppConfig {
  static const String _override = String.fromEnvironment('CHAOS_MASTER_URL');

  /// The origin (scheme + host + port) hosting the master.
  ///
  /// gRPC-Web method paths are absolute (they start with `/`), so a hosting
  /// sub-path cannot be preserved and only the origin is used.
  static String get masterUrl {
    if (_override.isNotEmpty) {
      return _override;
    }

    final base = web.document.baseURI;
    if (base.isEmpty) {
      return 'http://localhost:8080';
    }

    return Uri.parse(base).origin;
  }
}
