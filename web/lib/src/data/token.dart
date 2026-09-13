import 'package:grpc/grpc.dart';

/// Holds the bearer token shared by all gRPC-Web clients.
class TokenStore {
  String token = '';

  /// Invoked when the master rejects a call as unauthenticated so the UI can
  /// return the user to the login screen.
  void Function()? onUnauthorized;
}

/// Adds the current bearer token to every outgoing gRPC call.
class AuthInterceptor extends ClientInterceptor {
  final TokenStore _store;

  AuthInterceptor(this._store);

  CallOptions _withToken(CallOptions options) {
    final metadata = authMetadata(_store.token);
    if (metadata == null) {
      return options;
    }

    return options.mergedWith(CallOptions(metadata: metadata));
  }

  @override
  ResponseFuture<R> interceptUnary<Q, R>(
    ClientMethod<Q, R> method,
    Q request,
    CallOptions options,
    ClientUnaryInvoker<Q, R> invoker,
  ) {
    return invoker(method, request, _withToken(options));
  }

  @override
  ResponseStream<R> interceptStreaming<Q, R>(
    ClientMethod<Q, R> method,
    Stream<Q> requests,
    CallOptions options,
    ClientStreamingInvoker<Q, R> invoker,
  ) {
    return invoker(method, requests, _withToken(options));
  }
}

/// Builds the authorization metadata for [token], or null when it is empty.
Map<String, String>? authMetadata(String token) {
  if (token.isEmpty) {
    return null;
  }

  return {'authorization': 'Bearer $token'};
}
