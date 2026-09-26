import 'dart:convert';
import 'dart:js_interop';

import 'package:web/web.dart' as web;

/// Triggers a browser download of [contents] as [filename] with the given MIME
/// [mimeType]. This is the single place the web app creates downloads.
void downloadText(
  String filename,
  String contents, {
  String mimeType = 'text/plain;charset=utf-8',
}) {
  final bytes = utf8.encode(contents);
  final blob = web.Blob(
    [bytes.toJS].toJS,
    web.BlobPropertyBag(type: mimeType),
  );
  final url = web.URL.createObjectURL(blob);
  final anchor = web.HTMLAnchorElement()
    ..href = url
    ..download = filename;
  anchor.click();
  web.URL.revokeObjectURL(url);
}

/// Convenience wrapper that downloads pretty-printed JSON.
void downloadJson(String filename, String json) {
  downloadText(filename, json, mimeType: 'application/json');
}
