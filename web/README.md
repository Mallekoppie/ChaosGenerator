# Chaos Master web app

Flutter web control panel for the Chaos Generator master. It talks to the
master over **gRPC-Web** using the protobuf contract in `../internal/contracts`.

## Generated code

The Dart protobuf/gRPC-Web stubs are generated from the Go contracts and are
**not** checked in. Generate them before building:

```bash
make proto-dart
```

This requires:

- the Dart SDK (shipped with Flutter)
- `protoc` on the `PATH`
- the Dart protoc plugin:

```bash
dart pub global activate protoc_plugin
```

Generated files land in `lib/src/generated/`:

- `master.pb.dart` / `master.pbgrpc.dart`
- `auth.pb.dart` / `auth.pbgrpc.dart`

## Local development

Run the master with `DEBUG=true` (gRPC-Web only, no embedded assets):

```bash
make run-master-dev
```

Then run the app against it:

```bash
cd web
flutter pub get
flutter run -d chrome --dart-define=CHAOS_MASTER_URL=http://localhost:9001
```

When the app is served by the master itself the URL is derived from
`document.baseUri`, so no `--dart-define` is required.

## Build

```bash
make buildweb        # flutter build web --release --base-href=/  -> ../cmd/master/dist
```

## Tests

```bash
cd web
flutter test
```
