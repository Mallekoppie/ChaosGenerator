import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

import '../theme.dart';

/// One named series rendered by [AetherLineChart].
class AetherSeries {
  const AetherSeries({
    required this.label,
    required this.color,
    required this.points,
  });

  final String label;
  final Color color;
  final List<FlSpot> points;
}

/// Horizontal reference line, used for the latency SLA ceiling.
class AetherReferenceLine {
  const AetherReferenceLine({
    required this.value,
    required this.label,
    this.color = AetherPalette.amber,
  });

  final double value;
  final String label;
  final Color color;
}

/// Aether-themed multi-series line chart.
///
/// Keeps fl_chart's styling in one place so every telemetry panel shares the
/// same grid, axis and tooltip treatment. X values are expected to be dense
/// sample indices; pass [xLabelBuilder] to render real labels on the axis.
class AetherLineChart extends StatelessWidget {
  const AetherLineChart({
    required this.series,
    this.referenceLines = const [],
    this.minY = 0,
    this.maxY,
    this.yLabelSuffix = '',
    this.yInterval,
    this.xLabelBuilder,
    this.height = 210,
    this.showArea = true,
    super.key,
  });

  final List<AetherSeries> series;
  final List<AetherReferenceLine> referenceLines;
  final double minY;
  final double? maxY;
  final String yLabelSuffix;
  final double? yInterval;
  final String Function(double value)? xLabelBuilder;
  final double height;
  final bool showArea;

  @override
  Widget build(BuildContext context) {
    final visible = series.where((s) => s.points.isNotEmpty).toList();

    if (visible.isEmpty) {
      return SizedBox(
        height: height,
        child: const Center(
          child: Text(
            'Awaiting telemetry',
            style: TextStyle(color: AetherPalette.textMuted, fontSize: 12.5),
          ),
        ),
      );
    }

    final maxX = visible
        .expand((s) => s.points)
        .map((spot) => spot.x)
        .reduce((a, b) => a > b ? a : b);
    // fl_chart needs a non-zero span, which a single sample would not provide.
    final safeMaxX = maxX <= 0 ? 1.0 : maxX;

    return SizedBox(
      height: height,
      child: LineChart(
        LineChartData(
          minX: 0,
          maxX: safeMaxX,
          minY: minY,
          maxY: maxY,
          clipData: const FlClipData.all(),
          gridData: FlGridData(
            show: true,
            drawVerticalLine: false,
            horizontalInterval: yInterval,
            getDrawingHorizontalLine: (value) =>
                const FlLine(color: AetherPalette.borderSoft, strokeWidth: 1),
          ),
          borderData: FlBorderData(
            show: true,
            border: const Border(
              left: BorderSide(color: AetherPalette.borderSoft),
              bottom: BorderSide(color: AetherPalette.borderSoft),
            ),
          ),
          titlesData: FlTitlesData(
            topTitles: const AxisTitles(
              sideTitles: SideTitles(showTitles: false),
            ),
            rightTitles: const AxisTitles(
              sideTitles: SideTitles(showTitles: false),
            ),
            leftTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                reservedSize: 46,
                interval: yInterval,
                getTitlesWidget: (value, meta) => Padding(
                  padding: const EdgeInsets.only(right: 6),
                  child: Text(
                    '${_compact(value)}$yLabelSuffix',
                    style: AetherText.mono(
                      size: 10,
                      color: AetherPalette.textMuted,
                    ),
                    textAlign: TextAlign.right,
                  ),
                ),
              ),
            ),
            bottomTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: xLabelBuilder != null,
                reservedSize: 26,
                getTitlesWidget: (value, meta) => Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: Text(
                    xLabelBuilder?.call(value) ?? '',
                    style: AetherText.mono(
                      size: 10,
                      color: AetherPalette.textMuted,
                    ),
                  ),
                ),
              ),
            ),
          ),
          extraLinesData: ExtraLinesData(
            horizontalLines: [
              for (final line in referenceLines)
                HorizontalLine(
                  y: line.value,
                  color: line.color,
                  strokeWidth: 1.2,
                  dashArray: const [6, 4],
                  label: HorizontalLineLabel(
                    show: true,
                    alignment: Alignment.topRight,
                    style: AetherText.mono(
                      size: 9.5,
                      weight: FontWeight.w700,
                      color: line.color,
                    ),
                    labelResolver: (_) => line.label,
                  ),
                ),
            ],
          ),
          lineTouchData: LineTouchData(
            enabled: true,
            touchTooltipData: LineTouchTooltipData(
              getTooltipColor: (_) => AetherPalette.surfaceHighest,
              getTooltipItems: (spots) => [
                for (final spot in spots)
                  LineTooltipItem(
                    '${visible[spot.barIndex].label}: ${_compact(spot.y)}$yLabelSuffix',
                    AetherText.mono(
                      size: 11,
                      color: visible[spot.barIndex].color,
                    ),
                  ),
              ],
            ),
          ),
          lineBarsData: [
            for (final item in visible)
              LineChartBarData(
                spots: item.points,
                isCurved: true,
                curveSmoothness: 0.2,
                color: item.color,
                barWidth: 1.8,
                isStrokeCapRound: true,
                dotData: const FlDotData(show: false),
                belowBarData: BarAreaData(
                  show: showArea,
                  color: item.color.withValues(alpha: 0.08),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

/// Minimal inline sparkline for the KPI cards.
class AetherSparkline extends StatelessWidget {
  const AetherSparkline({
    required this.points,
    this.color = AetherPalette.cyanBright,
    this.height = 30,
    super.key,
  });

  final List<FlSpot> points;
  final Color color;
  final double height;

  @override
  Widget build(BuildContext context) {
    if (points.length < 2) {
      return SizedBox(height: height);
    }

    final maxX = points.map((spot) => spot.x).reduce((a, b) => a > b ? a : b);
    final maxY = points.map((spot) => spot.y).reduce((a, b) => a > b ? a : b);

    return SizedBox(
      height: height,
      child: LineChart(
        LineChartData(
          minX: 0,
          maxX: maxX,
          minY: 0,
          maxY: maxY <= 0 ? 1 : maxY,
          gridData: const FlGridData(show: false),
          borderData: FlBorderData(show: false),
          titlesData: const FlTitlesData(show: false),
          lineTouchData: const LineTouchData(enabled: false),
          lineBarsData: [
            LineChartBarData(
              spots: points,
              isCurved: true,
              curveSmoothness: 0.25,
              color: color,
              barWidth: 1.6,
              isStrokeCapRound: true,
              dotData: const FlDotData(show: false),
              belowBarData: BarAreaData(
                show: true,
                color: color.withValues(alpha: 0.12),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Legend row shared by the chart panels.
class AetherChartLegend extends StatelessWidget {
  const AetherChartLegend({required this.entries, super.key});

  final List<({String label, Color color})> entries;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 16,
      runSpacing: 6,
      children: [
        for (final entry in entries)
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 10,
                height: 3,
                decoration: BoxDecoration(
                  color: entry.color,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const SizedBox(width: 6),
              Text(entry.label.toUpperCase(), style: AetherText.sectionLabel),
            ],
          ),
      ],
    );
  }
}

/// Formats a value compactly for axes and tooltips (420000 -> 420k).
String _compact(double value) {
  final absolute = value.abs();

  if (absolute >= 1000000) {
    return '${(value / 1000000).toStringAsFixed(1)}M';
  }
  if (absolute >= 1000) {
    return '${(value / 1000).toStringAsFixed(1)}k';
  }
  if (absolute >= 100 || absolute == 0) {
    return value.toStringAsFixed(0);
  }
  return value.toStringAsFixed(1);
}
