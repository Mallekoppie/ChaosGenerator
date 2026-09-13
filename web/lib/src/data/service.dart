import 'package:grpc/grpc.dart';
import 'package:grpc/grpc_web.dart' show GrpcWebClientChannel;

import '../generated/generated.dart';
import 'config.dart';
import 'token.dart';

/// Thin wrapper around the generated gRPC-Web clients.
class Backend {
  final TokenStore tokenStore;

  late final GrpcWebClientChannel _channel;
  late final ChaosMasterClient master;
  late final AuthClient auth;

  Backend({required this.tokenStore}) {
    final interceptor = AuthInterceptor(tokenStore);

    _channel = GrpcWebClientChannel.xhr(Uri.parse(AppConfig.masterUrl));
    master = ChaosMasterClient(_channel, interceptors: [interceptor]);
    auth = AuthClient(_channel, interceptors: [interceptor]);
  }

  /// Notifies the token store when the master reports an unauthenticated call.
  void handleError(Object error) {
    if (error is GrpcError && error.code == StatusCode.unauthenticated) {
      tokenStore.onUnauthorized?.call();
    }
  }

  Future<void> shutdown() async {
    await _channel.shutdown();
  }
}
