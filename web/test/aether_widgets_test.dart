import 'dart:ui' as ui;

import 'package:chaos_master_web/src/theme.dart';
import 'package:chaos_master_web/src/widgets/aether_chrome.dart';
import 'package:chaos_master_web/src/widgets/aether_common.dart';
import 'package:chaos_master_web/src/widgets/aether_kpi_card.dart';
import 'package:chaos_master_web/src/widgets/aether_panel.dart';
import 'package:chaos_master_web/src/widgets/aether_table.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';

/// Hosts [child] in the app theme at a fixed viewport size, so layout
/// regressions (overflow, unbounded constraints) surface as test failures.
///
/// Both the `MediaQuery` and the layout box are pinned: `LayoutBuilder`-driven
/// widgets read the box, while `MediaQuery.sizeOf`-driven widgets read the
/// media query, and the two must agree for the responsive branches to be
/// exercised honestly.
Widget _host(Widget child, {double width = 1440, double height = 900}) {
  return MaterialApp(
    theme: aetherTheme(),
    home: MediaQuery(
      data: MediaQueryData(size: Size(width, height)),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        body: AetherBackground(
          child: SizedBox(width: width, height: height, child: child),
        ),
      ),
    ),
  );
}

Widget _table() {
  return AetherDataTable(
    columns: const [
      DataColumn(label: Text('ID')),
      DataColumn(label: Text('Host')),
      DataColumn(label: Text('Port')),
      DataColumn(label: Text('Enabled')),
      DataColumn(label: Text('Status')),
    ],
    rows: [
      DataRow(
        cells: [
          const DataCell(Text('agent-1')),
          const DataCell(Text('10.0.0.1')),
          const DataCell(Text('9003')),
          DataCell(boolCell(true)),
          DataCell(TelemetryText('online', color: statusColor('online'))),
        ],
      ),
    ],
  );
}

void main() {
  testWidgets('KPI deck lays out without overflow on a wide viewport', (
    tester,
  ) async {
    await tester.pumpWidget(
      _host(
        const KpiDeck(
          cards: [
            KpiCard(
              label: 'Total synthetic concurrency',
              value: '420,000',
              unit: 'req/s',
              hint: 'fleet-wide egress estimate',
            ),
            KpiCard(label: 'gRPC fleet ping', value: '1.82', unit: 'ms avg'),
            KpiCard(label: 'Chaos status', value: 'IDLE'),
          ],
        ),
      ),
    );

    expect(find.text('420,000'), findsOneWidget);
    expect(find.text('CHAOS STATUS'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('KPI deck wraps instead of overflowing on a narrow viewport', (
    tester,
  ) async {
    await tester.pumpWidget(
      _host(
        const KpiDeck(
          cards: [
            KpiCard(label: 'A', value: '1'),
            KpiCard(label: 'B', value: '2'),
            KpiCard(label: 'C', value: '3'),
          ],
        ),
        width: 700,
      ),
    );

    expect(tester.takeException(), isNull);
  });

  testWidgets('titled panel renders a table without overflow', (tester) async {
    await tester.pumpWidget(
      _host(
        AetherPanel(
          title: 'Agent fleet',
          padding: EdgeInsets.zero,
          child: _table(),
        ),
      ),
    );

    expect(find.text('AGENT FLEET'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('filling panel scrolls its child instead of overflowing', (
    tester,
  ) async {
    await tester.pumpWidget(
      _host(
        Column(
          children: [
            const SizedBox(height: 120),
            Expanded(
              child: AetherPanel(
                title: 'Running tests',
                fill: true,
                padding: EdgeInsets.zero,
                child: SingleChildScrollView(child: _table()),
              ),
            ),
          ],
        ),
        height: 500,
      ),
    );

    expect(find.text('RUNNING TESTS'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('table scrolls horizontally when narrower than its columns', (
    tester,
  ) async {
    await tester.pumpWidget(
      _host(
        AetherPanel(
          title: 'Registry',
          padding: EdgeInsets.zero,
          child: _table(),
        ),
        width: 420,
      ),
    );

    expect(tester.takeException(), isNull);
  });

  testWidgets('empty state message renders inside a panel', (tester) async {
    await tester.pumpWidget(
      _host(
        const AetherPanel(
          title: 'Agent fleet',
          child: ConsoleMessage('No agents are connected'),
        ),
      ),
    );

    expect(find.text('No agents are connected'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('global top bar keeps the brand and drops optional content '
      'on small viewports', (tester) async {
    await tester.pumpWidget(_host(const GlobalTopBar(), width: 600));

    expect(find.text('CHAOSPROCESSOR'), findsOneWidget);
    expect(find.text('GRID: NOMINAL'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('global top bar shows status pills on wide viewports', (
    tester,
  ) async {
    await tester.pumpWidget(_host(const GlobalTopBar()));

    expect(find.text('GRID: NOMINAL'), findsOneWidget);
    expect(find.text('CONTROLLED ENTROPY'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('secondary bar renders with ops pills and operator badge', (
    tester,
  ) async {
    await tester.pumpWidget(
      _host(
        Builder(
          builder: (context) => Scaffold(
            appBar: aetherAppBar(
              context,
              title: 'Agents',
              actions: [
                IconButton(onPressed: () {}, icon: const Icon(Icons.refresh)),
              ],
            ),
          ),
        ),
      ),
    );

    expect(find.text('AGENTS'), findsOneWidget);
    expect(find.text('LATENCY SLA: 99.95%'), findsOneWidget);
    expect(find.text('OPERATOR'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('method chips colour-code by HTTP verb', (tester) async {
    await tester.pumpWidget(
      _host(
        const Row(
          children: [
            MethodChip('GET'),
            MethodChip('POST'),
            MethodChip('PUT'),
            MethodChip('DELETE'),
          ],
        ),
      ),
    );

    expect(MethodChip.colorFor('get'), AetherPalette.cyanBright);
    expect(MethodChip.colorFor('post'), AetherPalette.emeraldBright);
    expect(MethodChip.colorFor('put'), AetherPalette.amber);
    expect(MethodChip.colorFor('delete'), AetherPalette.crimson);
    expect(tester.takeException(), isNull);
  });

  testWidgets('status and protocol colours are semantic', (tester) async {
    expect(statusColor('online'), AetherPalette.emeraldBright);
    expect(statusColor('degraded'), AetherPalette.amber);
    expect(statusColor('offline'), AetherPalette.textMuted);
    expect(protocolColor('https'), AetherPalette.emeraldBright);
    expect(protocolColor('grpc'), AetherPalette.amber);
    expect(protocolColor('http'), AetherPalette.cyanBright);
  });

  testWidgets('backdrop paints an opaque dark canvas instead of letting the '
      'transparent page background bleed through', (tester) async {
    final boundaryKey = GlobalKey();

    await tester.pumpWidget(
      RepaintBoundary(key: boundaryKey, child: _host(const SizedBox.expand())),
    );

    final boundary =
        boundaryKey.currentContext!.findRenderObject()!
            as RenderRepaintBoundary;

    await tester.runAsync(() async {
      final rendered = await boundary.toImage();
      final data = await rendered.toByteData(
        format: ui.ImageByteFormat.rawRgba,
      );
      final pixels = data!.buffer.asUint8List();
      final width = rendered.width;
      final maxX = rendered.width - 4;
      final maxY = rendered.height - 4;

      void expectOpaqueAndDark(int x, int y) {
        final offset = (y * width + x) * 4;
        final where = 'at ($x, $y)';

        expect(pixels[offset + 3], 255, reason: 'backdrop alpha $where');
        expect(pixels[offset], lessThan(80), reason: 'backdrop red $where');
        expect(
          pixels[offset + 1],
          lessThan(90),
          reason: 'backdrop green $where',
        );
        expect(
          pixels[offset + 2],
          lessThan(110),
          reason: 'backdrop blue $where',
        );
      }

      // The bloom is brightest top-left and the flat canvas darkest
      // bottom-right; every corner must stay dark *and* fully opaque.
      expectOpaqueAndDark(3, 3);
      expectOpaqueAndDark(3, maxY);
      expectOpaqueAndDark(maxX, 3);
      expectOpaqueAndDark(maxX, maxY);
    });
  });
}
