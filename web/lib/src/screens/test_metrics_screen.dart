import 'dart:async';
import 'dart:convert';
import 'dart:js_interop';
import 'dart:math' as math;

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:web/web.dart' as web;

import '../data/repo.dart';
import '../generated/generated.dart';
import '../theme.dart';
import '../widgets/aether_charts.dart';
import '../widgets/aether_chrome.dart';
import '../widgets/aether_common.dart';
import '../widgets/aether_kpi_card.dart';
import '../widgets/aether_panel.dart';
import '../widgets/aether_table.dart';

/// Latency SLA ceiling drawn on the envelope chart, in milliseconds.
const double _latencySlaCeilingMs = 500;

/// **Test metrics drilldown** — `/tests/metrics/:id`.
///
/// Visualises the cumulative reports agents push to the master over their gRPC
/// stream. Everything shown here is client-observed: the master never scrapes
/// the target or, in this path, the agents.
class TestMetricsScreen extends StatefulWidget {
  const TestMetricsScreen({
    required this.executionId,
    required this.metrics,
    required this.executions,
    super.key,
  });

  final String executionId;
  final TestMetricsRepository metrics;
  final ExecutionsRepository executions;

  @override
  State<TestMetricsScreen> createState() => _TestMetricsScreenState();
}

class _TestMetricsScreenState extends State<TestMetricsScreen> {
  Timer? _ticker;
  bool _aborting = false;

  @override
  void initState() {
    super.initState();

    widget.metrics.reset();
    widget.metrics.load(widget.executionId);
    widget.metrics.startPolling(widget.executionId);

    // Drives the elapsed timer once a second without waiting for a poll.
    _ticker = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) {
        setState(() {});
      }
    });
  }

  @override
  void dispose() {
    _ticker?.cancel();
    widget.metrics.stopPolling();
    super.dispose();
  }

  void _showMessage(String message) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _abort() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Abort run?'),
        content: Text(
          'Stop test execution ${widget.executionId} on every agent?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: AetherPalette.crimsonDeep,
              foregroundColor: Colors.white,
            ),
            onPressed: () => Navigator.of(dialogContext).pop(true),
            child: const Text('Abort run'),
          ),
        ],
      ),
    );

    if (confirmed != true) {
      return;
    }

    setState(() => _aborting = true);
    final error = await widget.executions.stop(widget.executionId);

    if (!mounted) {
      return;
    }

    setState(() => _aborting = false);
    _showMessage(error ?? 'Abort dispatched');
  }

  void _export(GetTestMetricsResponse response) {
    final summary = response.summary;

    final payload = <String, Object?>{
      'executionId': summary.testExecutionId,
      'useCase': {'id': summary.useCaseId, 'name': summary.useCaseName},
      'target': {
        'id': summary.targetId,
        'name': summary.targetName,
        'address': summary.targetAddress,
        'protocol': summary.targetProtocol,
      },
      'summary': {
        'running': summary.running,
        'startTime': summary.startTime.toInt(),
        'egressRps': summary.egressRps,
        'p50Ms': summary.p50Ms,
        'p90Ms': summary.p90Ms,
        'p95Ms': summary.p95Ms,
        'p99Ms': summary.p99Ms,
        'errorRate': summary.errorRate,
        'totalRequests': summary.totalRequests.toInt(),
        'totalErrors': summary.totalErrors.toInt(),
        'successRequests': summary.successRequests.toInt(),
        'clientErrors': summary.clientErrors.toInt(),
        'serverErrors': summary.serverErrors.toInt(),
        'timeouts': summary.timeouts.toInt(),
        'connectionResets': summary.connectionResets.toInt(),
        'errorsByCode': summary.errorsByCode.map(
          (key, value) => MapEntry(key, value.toInt()),
        ),
        'activeUsers': summary.activeUsers.toInt(),
        'connectedAgents': summary.connectedAgents,
        'numberOfAgents': summary.numberOfAgents,
      },
      'samples': [
        for (final sample in response.samples)
          {
            'timestampMs': sample.timestampMs.toInt(),
            'egressRps': sample.egressRps,
            'p50Ms': sample.p50Ms,
            'p90Ms': sample.p90Ms,
            'p95Ms': sample.p95Ms,
            'p99Ms': sample.p99Ms,
            'ok': sample.ok.toInt(),
            'clientErrors': sample.clientErrors.toInt(),
            'serverErrors': sample.serverErrors.toInt(),
            'timeouts': sample.timeouts.toInt(),
            'resets': sample.resets.toInt(),
            'errorRate': sample.errorRate,
          },
      ],
      'workers': [
        for (final worker in response.workers)
          {
            'agentId': worker.agentId,
            'online': worker.online,
            'activeUsers': worker.activeUsers,
            'egressRps': worker.egressRps,
            'observedP99Ms': worker.observedP99Ms,
            'cpuPercent': worker.cpuPercent,
            'memoryBytes': worker.memoryBytes.toInt(),
            'metricsScrapes': worker.metricsScrapes.toInt(),
            'lastUpdateMs': worker.lastUpdateMs.toInt(),
            'errorRate': worker.errorRate,
          },
      ],
    };

    final json = const JsonEncoder.withIndent('  ').convert(payload);
    _downloadJson('${widget.executionId}-metrics.json', json);
    _showMessage('Metrics dump exported');
  }

  String _elapsed(TestMetricsSummary summary) {
    final start = summary.startTime.toInt();
    if (start <= 0) {
      return '00:00:00';
    }

    final started = DateTime.fromMillisecondsSinceEpoch(start * 1000);
    var seconds = DateTime.now().difference(started).inSeconds;
    if (seconds < 0) {
      seconds = 0;
    }

    final hours = (seconds ~/ 3600).toString().padLeft(2, '0');
    final minutes = ((seconds % 3600) ~/ 60).toString().padLeft(2, '0');
    final secs = (seconds % 60).toString().padLeft(2, '0');
    return '$hours:$minutes:$secs';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: aetherAppBar(
        context,
        title: 'Test metrics',
        actions: [
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: () => widget.metrics.load(widget.executionId),
          ),
        ],
      ),
      body: ListenableBuilder(
        listenable: widget.metrics,
        builder: (context, _) {
          final repository = widget.metrics;
          final response = repository.response;

          if (response == null) {
            if (repository.loading) {
              return const ConsoleLoading();
            }

            return ConsoleMessage(
              repository.error ??
                  'No metrics available for this test execution',
              icon: Icons.query_stats,
              color: repository.error != null ? AetherPalette.crimson : null,
            );
          }

          final summary = response.summary;

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              _ControlBar(
                summary: summary,
                elapsed: _elapsed(summary),
                aborting: _aborting,
                onExport: () => _export(response),
                onAbort: _abort,
              ),
              const SizedBox(height: 16),
              const _PerspectiveBanner(),
              const SizedBox(height: 16),
              _buildKpis(summary, response.samples),
              const SizedBox(height: 16),
              _buildRequestRateChart(response.samples),
              const SizedBox(height: 16),
              _buildLatencyChart(response.samples),
              const SizedBox(height: 16),
              _buildOperationalRow(response.workers),
              const SizedBox(height: 16),
              _buildWorkers(response.workers),
            ],
          );
        },
      ),
    );
  }

  Widget _buildKpis(
    TestMetricsSummary summary,
    List<TestMetricsSample> samples,
  ) {
    final egressSpots = _spots(samples, (sample) => sample.egressRps);

    return KpiDeck(
      cards: [
        KpiCard(
          label: 'Client generated load',
          value: _formatRate(summary.egressRps),
          unit: 'req/s',
          icon: Icons.bolt,
          accent: AetherPalette.cyanGlow,
          hint: 'outbound synthetic egress',
          sparkline: egressSpots.length >= 2
              ? AetherSparkline(
                  points: egressSpots,
                  color: AetherPalette.cyanGlow,
                )
              : null,
        ),
        KpiCard(
          label: 'Agent round-trip latency',
          value: _formatMs(summary.p95Ms),
          unit: 'ms p95',
          icon: Icons.timer_outlined,
          accent: AetherPalette.cyanBright,
          hint:
              'p50 ${_formatMs(summary.p50Ms)} · p90 ${_formatMs(summary.p90Ms)} · p99 ${_formatMs(summary.p99Ms)}',
        ),
        KpiCard(
          label: 'Client-observed failures',
          value: _formatPercent(summary.errorRate),
          unit: 'errors',
          icon: Icons.warning_amber,
          accent: summary.totalErrors == 0
              ? AetherPalette.emeraldBright
              : AetherPalette.crimson,
          hint:
              '5xx ${summary.serverErrors} · 4xx ${summary.clientErrors} · timeouts ${summary.timeouts}',
        ),
        KpiCard(
          label: 'Distributed node fleet',
          value: '${summary.connectedAgents}/${summary.numberOfAgents}',
          unit: 'online',
          icon: Icons.dns,
          accent: AetherPalette.emeraldBright,
          hint: '${summary.activeUsers} virtual users',
        ),
      ],
    );
  }

  Widget _buildRequestRateChart(List<TestMetricsSample> samples) {
    final series = [
      AetherSeries(
        label: 'OK',
        color: AetherPalette.emeraldBright,
        points: _spots(samples, (sample) => sample.ok.toDouble()),
      ),
      AetherSeries(
        label: '4xx',
        color: AetherPalette.amber,
        points: _spots(samples, (sample) => sample.clientErrors.toDouble()),
      ),
      AetherSeries(
        label: '5xx',
        color: AetherPalette.crimson,
        points: _spots(samples, (sample) => sample.serverErrors.toDouble()),
      ),
      AetherSeries(
        label: 'Timeouts',
        color: AetherPalette.cyanBright,
        points: _spots(samples, (sample) => sample.timeouts.toDouble()),
      ),
    ];

    final maxY = _maxY(series);

    return AetherPanel(
      title: 'Client request rate & status demux',
      accent: AetherPalette.cyanBright,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          AetherChartLegend(
            entries: [
              for (final item in series) (label: item.label, color: item.color),
            ],
          ),
          const SizedBox(height: 18),
          AetherLineChart(
            series: series,
            maxY: maxY,
            yInterval: _intervalFor(maxY),
            xLabelBuilder: _timeLabel,
            height: 220,
          ),
        ],
      ),
    );
  }

  Widget _buildLatencyChart(List<TestMetricsSample> samples) {
    final series = [
      AetherSeries(
        label: 'p50',
        color: AetherPalette.cyanBright,
        points: _spots(samples, (sample) => sample.p50Ms),
      ),
      AetherSeries(
        label: 'p90',
        color: AetherPalette.cyan,
        points: _spots(samples, (sample) => sample.p90Ms),
      ),
      AetherSeries(
        label: 'p95',
        color: AetherPalette.amber,
        points: _spots(samples, (sample) => sample.p95Ms),
      ),
      AetherSeries(
        label: 'p99',
        color: AetherPalette.crimson,
        points: _spots(samples, (sample) => sample.p99Ms),
      ),
    ];

    final maxY = math.max(_maxY(series), _latencySlaCeilingMs * 1.25);
    final slaLabel = 'SLA ${_latencySlaCeilingMs.toStringAsFixed(0)}ms';

    return AetherPanel(
      title: 'Round-trip latency envelope',
      accent: AetherPalette.amber,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          AetherChartLegend(
            entries: [
              for (final item in series) (label: item.label, color: item.color),
              (label: slaLabel, color: AetherPalette.amber),
            ],
          ),
          const SizedBox(height: 18),
          AetherLineChart(
            series: series,
            referenceLines: [
              AetherReferenceLine(value: _latencySlaCeilingMs, label: slaLabel),
            ],
            maxY: maxY,
            yInterval: _intervalFor(maxY),
            yLabelSuffix: 'ms',
            xLabelBuilder: _timeLabel,
            height: 220,
            showArea: false,
          ),
        ],
      ),
    );
  }

  Widget _buildOperationalRow(List<TestMetricsWorker> workers) {
    return LayoutBuilder(
      builder: (context, constraints) {
        if (constraints.maxWidth < 1100) {
          return Column(
            children: [
              const _PerturbationsPanel(),
              const SizedBox(height: 16),
              _ScraperHealthPanel(workers: workers),
            ],
          );
        }

        return Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Expanded(child: _PerturbationsPanel()),
            const SizedBox(width: 16),
            Expanded(child: _ScraperHealthPanel(workers: workers)),
          ],
        );
      },
    );
  }

  Widget _buildWorkers(List<TestMetricsWorker> workers) {
    return AetherPanel(
      title: 'Agent fleet runtime',
      padding: EdgeInsets.zero,
      child: workers.isEmpty
          ? const ConsoleMessage(
              'No worker telemetry yet',
              icon: Icons.dns_outlined,
            )
          : AetherDataTable(
              columns: const [
                DataColumn(label: Text('Agent')),
                DataColumn(label: Text('Status')),
                DataColumn(label: Text('V-users')),
                DataColumn(label: Text('Egress RPS')),
                DataColumn(label: Text('Observed p99')),
                DataColumn(label: Text('Node CPU')),
                DataColumn(label: Text('Node memory')),
                DataColumn(label: Text('Prometheus')),
              ],
              rows: [
                for (final worker in workers)
                  DataRow(
                    color: AetherDataTable.rowHighlight,
                    cells: [
                      DataCell(
                        TelemetryText(
                          worker.agentId,
                          size: 12,
                          color: AetherPalette.cyanBright,
                        ),
                      ),
                      DataCell(_workerStatus(worker)),
                      DataCell(TelemetryText('${worker.activeUsers}')),
                      DataCell(TelemetryText(_formatRate(worker.egressRps))),
                      DataCell(
                        TelemetryText('${_formatMs(worker.observedP99Ms)} ms'),
                      ),
                      DataCell(_cpuCell(worker.cpuPercent)),
                      DataCell(
                        TelemetryText(_formatBytes(worker.memoryBytes.toInt())),
                      ),
                      DataCell(_prometheusCell(worker)),
                    ],
                  ),
              ],
            ),
    );
  }

  Widget _workerStatus(TestMetricsWorker worker) {
    final label = worker.online ? 'online' : 'offline';

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        StatusDot(
          color: worker.online
              ? AetherPalette.emeraldBright
              : AetherPalette.textMuted,
          size: 7,
        ),
        const SizedBox(width: 8),
        Text(
          label,
          style: TextStyle(color: statusColor(label), fontSize: 12.5),
        ),
      ],
    );
  }

  Widget _cpuCell(double cpuPercent) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox(
          width: 90,
          child: MeterBar(
            fraction: cpuPercent / 100,
            color: cpuPercent >= 85
                ? AetherPalette.crimson
                : AetherPalette.emeraldBright,
          ),
        ),
        const SizedBox(width: 10),
        TelemetryText(
          '${cpuPercent.toStringAsFixed(1)}%',
          size: 11.5,
          color: AetherPalette.textMuted,
        ),
      ],
    );
  }

  Widget _prometheusCell(TestMetricsWorker worker) {
    final scrapes = worker.metricsScrapes.toInt();

    if (scrapes == 0) {
      return TelemetryText(
        'gRPC only',
        size: 11.5,
        color: AetherPalette.textMuted,
      );
    }

    return TelemetryText(
      '$scrapes scrapes',
      size: 11.5,
      color: AetherPalette.emeraldBright,
    );
  }

  List<FlSpot> _spots(
    List<TestMetricsSample> samples,
    double Function(TestMetricsSample) value,
  ) {
    return [
      for (var i = 0; i < samples.length; i++)
        FlSpot(i.toDouble(), value(samples[i])),
    ];
  }

  double _maxY(List<AetherSeries> series) {
    var maxValue = 0.0;

    for (final item in series) {
      for (final spot in item.points) {
        if (spot.y > maxValue) {
          maxValue = spot.y;
        }
      }
    }

    return maxValue <= 0 ? 1 : maxValue * 1.15;
  }

  double _intervalFor(double maxY) {
    if (maxY <= 5) return 1;
    if (maxY <= 20) return 5;
    if (maxY <= 100) return 25;
    if (maxY <= 500) return 100;
    if (maxY <= 2000) return 500;
    return 1000;
  }

  String _timeLabel(double value) => 't+${value.toInt()}s';
}

class _ControlBar extends StatelessWidget {
  const _ControlBar({
    required this.summary,
    required this.elapsed,
    required this.aborting,
    required this.onExport,
    required this.onAbort,
  });

  final TestMetricsSummary summary;
  final String elapsed;
  final bool aborting;
  final VoidCallback onExport;
  final VoidCallback onAbort;

  @override
  Widget build(BuildContext context) {
    return AetherPanel(
      title: 'Run control',
      trailing: const StatusPill(
        label: 'Auto-scrape: 3s',
        color: AetherPalette.emeraldBright,
      ),
      child: Wrap(
        spacing: 24,
        runSpacing: 16,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: [
          _Field(label: 'Execution', value: summary.testExecutionId),
          _Field(
            label: 'Target',
            value: summary.targetAddress.isEmpty
                ? summary.targetName
                : summary.targetAddress,
          ),
          _Field(
            label: 'Use case',
            value: summary.useCaseName.isEmpty
                ? summary.useCaseId
                : summary.useCaseName,
          ),
          _Field(label: 'Elapsed', value: elapsed),
          _Field(
            label: 'Protocol',
            value: summary.targetProtocol.toUpperCase(),
          ),
          StatusPill(
            label: summary.running ? 'Running' : 'Stopped',
            color: summary.running
                ? AetherPalette.emeraldBright
                : AetherPalette.textMuted,
          ),
          OutlinedButton.icon(
            onPressed: onExport,
            icon: const Icon(Icons.download, size: 18),
            label: const Text('Export dump'),
          ),
          FilledButton.icon(
            onPressed: aborting ? null : onAbort,
            style: FilledButton.styleFrom(
              backgroundColor: AetherPalette.crimsonDeep,
              foregroundColor: Colors.white,
            ),
            icon: Icon(
              aborting ? Icons.hourglass_top : Icons.stop_circle_outlined,
              size: 18,
            ),
            label: Text(aborting ? 'Aborting...' : 'Abort run'),
          ),
        ],
      ),
    );
  }
}

class _Field extends StatelessWidget {
  const _Field({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(label.toUpperCase(), style: AetherText.sectionLabel),
        const SizedBox(height: 6),
        TelemetryText(
          value.isEmpty ? '—' : value,
          size: 12.5,
          color: AetherPalette.cyanBright,
        ),
      ],
    );
  }
}

class _PerspectiveBanner extends StatelessWidget {
  const _PerspectiveBanner();

  @override
  Widget build(BuildContext context) {
    return AetherPanel(
      title: 'Telemetry perspective',
      accent: AetherPalette.amber,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.info_outline, size: 18, color: AetherPalette.amber),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              'All measurements originate from the outbound synthetic agents. '
              'Targets sit outside the agent cluster, so no server-side target '
              'metrics are scraped — latency, throughput and failures are '
              'client-observed.',
              style: const TextStyle(
                color: AetherPalette.textMuted,
                fontSize: 12.5,
                height: 1.5,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Read-only placeholder until fault injection ships in the agents.
class _PerturbationsPanel extends StatelessWidget {
  const _PerturbationsPanel();

  @override
  Widget build(BuildContext context) {
    return AetherPanel(
      title: 'Active chaos perturbations',
      accent: AetherPalette.amber,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const StatusDot(color: AetherPalette.textMuted, size: 8),
              const SizedBox(width: 10),
              const Expanded(
                child: Text(
                  'No active perturbations',
                  style: TextStyle(
                    color: AetherPalette.textPrimary,
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
              const StatusPill(
                label: 'Not enabled',
                color: AetherPalette.textMuted,
              ),
            ],
          ),
          const SizedBox(height: 14),
          const Text(
            'Latency tail injection, socket jitter and TCP RST emulation are not '
            'wired into this build. Agents currently generate pure synthetic '
            'load and report the resulting client-observed telemetry.',
            style: TextStyle(
              color: AetherPalette.textMuted,
              fontSize: 12,
              height: 1.5,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScraperHealthPanel extends StatelessWidget {
  const _ScraperHealthPanel({required this.workers});

  final List<TestMetricsWorker> workers;

  @override
  Widget build(BuildContext context) {
    if (workers.isEmpty) {
      return const AetherPanel(
        title: 'Prometheus scraper health',
        child: ConsoleMessage(
          'No workers reporting',
          icon: Icons.monitor_heart_outlined,
        ),
      );
    }

    final totalScrapes = workers.fold<int>(
      0,
      (sum, worker) => sum + worker.metricsScrapes.toInt(),
    );
    final totalMemory = workers.fold<int>(
      0,
      (sum, worker) => sum + worker.memoryBytes.toInt(),
    );
    final freshest = workers
        .map((worker) => worker.lastUpdateMs.toInt())
        .reduce(math.max);
    final lagMs = DateTime.now().millisecondsSinceEpoch - freshest;

    return AetherPanel(
      title: 'Prometheus scraper health',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _healthRow('Scrapes observed', '$totalScrapes'),
          _healthRow(
            'Ingest lag',
            lagMs >= 0 ? '${(lagMs / 1000).toStringAsFixed(1)} s' : '—',
          ),
          _healthRow('Reporting workers', '${workers.length}'),
          _healthRow('Memory footprint', _formatBytes(totalMemory)),
          const SizedBox(height: 12),
          Text(
            totalScrapes == 0
                ? 'No Prometheus scrapes detected — telemetry is flowing over '
                      'gRPC only.'
                : 'Prometheus is scraping the agent fleet in parallel with the '
                      'gRPC telemetry feed.',
            style: const TextStyle(
              color: AetherPalette.textMuted,
              fontSize: 12,
              height: 1.5,
            ),
          ),
        ],
      ),
    );
  }

  Widget _healthRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: Text(label.toUpperCase(), style: AetherText.sectionLabel),
          ),
          TelemetryText(value, size: 12.5, color: AetherPalette.cyanBright),
        ],
      ),
    );
  }
}

String _formatRate(double value) {
  if (value >= 1000000) {
    return '${(value / 1000000).toStringAsFixed(2)}M';
  }
  if (value >= 10000) {
    return '${(value / 1000).toStringAsFixed(1)}k';
  }
  if (value >= 100) {
    return value.toStringAsFixed(0);
  }
  return value.toStringAsFixed(1);
}

String _formatMs(double value) {
  if (value <= 0) {
    return '0';
  }
  return value >= 100 ? value.toStringAsFixed(0) : value.toStringAsFixed(1);
}

String _formatPercent(double value) => '${(value * 100).toStringAsFixed(2)}%';

String _formatBytes(int bytes) {
  final value = bytes.toDouble();

  if (value >= 1024 * 1024 * 1024) {
    return '${(value / (1024 * 1024 * 1024)).toStringAsFixed(1)} GiB';
  }
  if (value >= 1024 * 1024) {
    return '${(value / (1024 * 1024)).toStringAsFixed(0)} MiB';
  }
  if (value >= 1024) {
    return '${(value / 1024).toStringAsFixed(0)} KiB';
  }
  return '${value.toStringAsFixed(0)} B';
}

void _downloadJson(String filename, String contents) {
  final bytes = utf8.encode(contents);
  final blob = web.Blob(
    [bytes.toJS].toJS,
    web.BlobPropertyBag(type: 'application/json'),
  );
  final url = web.URL.createObjectURL(blob);

  final anchor = web.HTMLAnchorElement()
    ..href = url
    ..download = filename;

  anchor.click();
  web.URL.revokeObjectURL(url);
}
