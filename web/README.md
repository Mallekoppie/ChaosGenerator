# ChaosProcessor web app

Flutter web control panel for the Chaos Generator master. It talks to the
master over **gRPC-Web** using the protobuf contract in `../internal/contracts`.

The UI uses the **Aether Flight Deck** ("Deep Space Cyan") dark console theme,
defined in `lib/src/theme.dart`. `STITCH-UI-BRIEF.md` documents the original
layout and `STICH-UI-RESPONSE.md` is the design spec it was rebuilt against.

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

The master runs in the kind cluster, so forward its port and point the app at it:

```bash
make proxy                                        # web UI + gRPC-Web on 9001
cd web
flutter pub get
flutter run -d chrome --dart-define=CHAOS_MASTER_URL=http://localhost:9001
```

When the app is served by the master itself (<http://localhost:9001/>) the URL is
derived from `document.baseUri`, so no `--dart-define` is required.

To get a rebuilt web app into the cluster, rebuild the images and roll the
deployments:

```bash
make kind-images kind-restart
```

## Build

```bash
make buildweb        # flutter build web --release --base-href=/  -> ../cmd/master/dist
```

## Tests

```bash
cd web
flutter test
```
