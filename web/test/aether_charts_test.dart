import 'package:chaos_master_web/src/theme.dart';
import 'package:chaos_master_web/src/widgets/aether_charts.dart';
import 'package:chaos_master_web/src/widgets/aether_kpi_card.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

/// Hosts [child] in the app theme at a fixed viewport size, so layout
/// regressions (overflow, unbounded constraints) surface as test failures.
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

AetherSeries _series(String label, Color color, List<double> values) {
  return AetherSeries(
    label: label,
    color: color,
    points: [
      for (var i = 0; i < values.length; i++) FlSpot(i.toDouble(), values[i]),
    ],
  );
}

void main() {
  testWidgets('line chart shows an empty state when no series has points', (
    tester,
  ) async {
    await tester.pumpWidget(
      _host(
        const AetherLineChart(
          series: [AetherSeries(label: 'OK', color: Colors.green, points: [])],
        ),
      ),
    );

    expect(find.text('Awaiting telemetry'), findsOneWidget);
    expect(find.byType(LineChart), findsNothing);
  });

  testWidgets('line chart renders multiple series and a reference line', (
    tester,
  ) async {
    await tester.pumpWidget(
      _host(
        AetherLineChart(
          series: [
            _series('OK', AetherPalette.emeraldBright, [10, 20, 30, 25]),
            _series('5xx', AetherPalette.crimson, [0, 1, 0, 2]),
          ],
          referenceLines: const [
            AetherReferenceLine(value: 25, label: 'SLA 25ms'),
          ],
          maxY: 40,
          yInterval: 10,
        ),
      ),
    );

    expect(tester.takeException(), isNull);
    expect(find.byType(LineChart), findsOneWidget);
  });

  testWidgets('sparkline stays empty for a single point', (tester) async {
    await tester.pumpWidget(
      _host(const AetherSparkline(points: [FlSpot(0, 1)])),
    );

    expect(find.byType(LineChart), findsNothing);
  });

  testWidgets('line chart tolerates a single sample', (tester) async {
    await tester.pumpWidget(
      _host(
        AetherLineChart(
          series: [
            _series('OK', AetherPalette.cyanBright, [12]),
          ],
        ),
      ),
    );

    expect(tester.takeException(), isNull);
    expect(find.byType(LineChart), findsOneWidget);
  });

  testWidgets('KPI card lays out with an embedded sparkline', (tester) async {
    await tester.pumpWidget(
      _host(
        KpiCard(
          label: 'Client generated load',
          value: '1.2k',
          unit: 'req/s',
          hint: 'outbound synthetic egress',
          sparkline: AetherSparkline(
            points: [for (var i = 0; i < 8; i++) FlSpot(i.toDouble(), i * 1.5)],
          ),
        ),
      ),
    );

    expect(tester.takeException(), isNull);
    expect(find.text('CLIENT GENERATED LOAD'), findsOneWidget);
    expect(find.byType(LineChart), findsOneWidget);
  });

  testWidgets('chart legend renders every entry', (tester) async {
    await tester.pumpWidget(
      _host(
        const AetherChartLegend(
          entries: [
            (label: 'p50', color: AetherPalette.cyanBright),
            (label: 'p99', color: AetherPalette.crimson),
          ],
        ),
      ),
    );

    expect(find.text('P50'), findsOneWidget);
    expect(find.text('P99'), findsOneWidget);
  });
}
