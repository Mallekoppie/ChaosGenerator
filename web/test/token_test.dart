import 'package:chaos_master_web/src/data/token.dart';
import 'package:test/test.dart';

void main() {
  test('authMetadata returns null when there is no token', () {
    expect(authMetadata(''), isNull);
  });

  test('authMetadata adds a bearer token', () {
    final metadata = authMetadata('abc123');

    expect(metadata, isNotNull);
    expect(metadata!['authorization'], 'Bearer abc123');
  });

  test('TokenStore starts empty and can be updated', () {
    final store = TokenStore();

    expect(store.token, isEmpty);

    store.token = 'a-token';

    expect(store.token, 'a-token');
  });
}
