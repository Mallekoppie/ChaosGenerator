import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

import '../data/repo.dart';
import '../generated/generated.dart';
import '../theme.dart';
import '../widgets/aether_charts.dart';
import '../widgets/aether_chrome.dart';
import '../widgets/aether_common.dart';
import '../widgets/aether_kpi_card.dart';
import '../widgets/aether_panel.dart';
import '../widgets/aether_table.dart';

/// Cadence at which the resource telemetry is polled. Kept in sync with the
/// repository default and used to label the sparkline trend window.
const _serverPollInterval = Duration(seconds: 3);

/// **SYS-CM** — executive telemetry overview.
///
/// Aggregates the data the control plane already exposes (registered agents,
/// targets, scenarios and live executions) into a single flight-deck view, and
/// surfaces the resource footprint of the master process and the agent fleet so
/// memory can be watched as a run scales out. Live per-socket telemetry (req/s,
/// latency envelopes, chaos injections) is not available yet and is deliberately
/// not shown here.
class TelemetryScreen extends StatefulWidget {
  const TelemetryScreen({
    required this.agents,
    required this.targets,
    required this.useCases,
    required this.executions,
    required this.telemetry,
    super.key,
  });

  final AgentsRepository agents;
  final TargetsRepository targets;
  final UseCasesRepository useCases;
  final ExecutionsRepository executions;
  final SystemTelemetryRepository telemetry;

  @override
  State<TelemetryScreen> createState() => _TelemetryScreenState();
}

class _TelemetryScreenState extends State<TelemetryScreen> {
  @override
  void initState() {
    super.initState();

    widget.agents.load();
    widget.targets.load();
    widget.useCases.load();
    widget.executions.load();
    widget.executions.startPolling();

    // Resource telemetry polls on the same cadence as the execution list so the
    // memory footprint tracks a run as it scales out.
    widget.telemetry.load();
    widget.telemetry.startPolling(interval: _serverPollInterval);
  }

  @override
  void dispose() {
    widget.executions.stopPolling();
    widget.telemetry.stopPolling();
    super.dispose();
  }

  Future<void> _refresh() async {
    await Future.wait([
      widget.agents.load(),
      widget.targets.load(),
      widget.useCases.load(),
      widget.executions.load(),
      widget.telemetry.load(),
    ]);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: aetherAppBar(
        context,
        title: 'SYS-CM',
        actions: [
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: _refresh,
          ),
        ],
      ),
      body: ListenableBuilder(
        listenable: Listenable.merge([
          widget.agents,
          widget.targets,
          widget.useCases,
          widget.executions,
          widget.telemetry,
        ]),
        builder: (context, _) {
          final agents = widget.agents.agents;
          final targets = widget.targets.targets;
          final useCases = widget.useCases.useCases;
          final executions = widget.executions.executions;

          final loading =
              widget.agents.loading &&
              widget.targets.loading &&
              widget.useCases.loading &&
              widget.executions.loading;
          final isEmpty =
              agents.isEmpty &&
              targets.isEmpty &&
              useCases.isEmpty &&
              executions.isEmpty;

          if (loading && isEmpty) {
            return const ConsoleLoading();
          }

          final error =
              widget.agents.error ??
              widget.targets.error ??
              widget.useCases.error ??
              widget.executions.error;

          if (error != null && isEmpty) {
            return ConsoleMessage(
              error,
              icon: Icons.error_outline,
              color: AetherPalette.crimson,
            );
          }

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              _buildKpiDeck(agents, targets, useCases, executions),
              const SizedBox(height: 16),
              _buildServerResources(),
              const SizedBox(height: 16),
              _buildAgentResources(),
              const SizedBox(height: 16),
              _buildRegistryBreakdown(agents, targets),
              const SizedBox(height: 16),
              _buildLiveDispatch(executions),
              const SizedBox(height: 16),
              const _TelemetryPerspectiveNote(),
            ],
          );
        },
      ),
    );
  }

  Widget _buildKpiDeck(
    List<Agent> agents,
    List<Target> targets,
    List<UseCase> useCases,
    List<RunningTestExecution> executions,
  ) {
    final enabledAgents = agents.where((agent) => agent.enabled).length;
    final virtualUsers = executions.fold<int>(
      0,
      (sum, execution) =>
          sum + execution.numberOfAgents * execution.simulatedUsersPerAgent,
    );

    return KpiDeck(
      cards: [
        KpiCard(
          label: 'Fleet registry',
          value: '${agents.length}',
          unit: 'agents',
          icon: Icons.dns,
          hint:
              '$enabledAgents enabled · ${agents.length - enabledAgents} disabled',
        ),
        KpiCard(
          label: 'Target registry',
          value: '${targets.length}',
          unit: 'targets',
          icon: Icons.my_location,
          accent: AetherPalette.cyanBright,
          hint: _protocolSplit(targets),
        ),
        KpiCard(
          label: 'Scenario catalogue',
          value: '${useCases.length}',
          unit: 'use cases',
          icon: Icons.list_alt,
          accent: AetherPalette.emeraldBright,
          hint: 'client-perspective probes',
        ),
        KpiCard(
          label: 'Active runs',
          value: '${executions.length}',
          unit: 'executions',
          icon: Icons.play_circle,
          accent: executions.isEmpty
              ? AetherPalette.emeraldBright
              : AetherPalette.amber,
          hint: 'auto-refresh every 3s',
        ),
        KpiCard(
          label: 'Synthetic concurrency',
          value: '$virtualUsers',
          unit: 'v-users',
          icon: Icons.bolt,
          accent: AetherPalette.cyanGlow,
          hint: 'fleet-wide virtual users',
        ),
      ],
    );
  }

  /// Control-plane resource footprint. Shown ahead of the agent fleet so the
  /// master's own CPU and memory stay visible when a run scales out.
  Widget _buildServerResources() {
    final telemetry = widget.telemetry;
    final server = telemetry.server;

    if (server == null) {
      return AetherPanel(
        title: 'Server resources',
        accent: AetherPalette.cyanBright,
        child: ConsoleMessage(
          telemetry.error ?? 'Waiting for the control plane to report',
          icon: Icons.dns_outlined,
          color: telemetry.error == null
              ? AetherPalette.textMuted
              : AetherPalette.crimson,
        ),
      );
    }

    final cpuSpots = _serverSpots(
      telemetry.serverHistory,
      (sample) => sample.cpuPercent,
    );
    final memorySpots = _serverSpots(
      telemetry.serverHistory,
      (sample) => sample.memoryBytes.toDouble(),
    );

    return AetherPanel(
      title: 'Server resources',
      accent: AetherPalette.cyanBright,
      trailing: const StatusPill(
        label: 'control plane',
        color: AetherPalette.cyanBright,
      ),
      child: KpiDeck(
        cards: [
          KpiCard(
            label: 'Server CPU',
            value: server.cpuPercent.toStringAsFixed(1),
            unit: '%',
            icon: Icons.speed,
            accent: _loadAccent(server.cpuPercent),
            hint:
                'master process · ${_windowLabel(telemetry.serverHistory.length)} trend',
            sparkline: cpuSpots.length >= 2
                ? AetherSparkline(
                    points: cpuSpots,
                    color: _loadAccent(server.cpuPercent),
                  )
                : null,
          ),
          KpiCard(
            label: 'Server memory',
            value: _formatBytes(server.memoryBytes.toInt()),
            unit: 'rss',
            icon: Icons.memory,
            accent: AetherPalette.cyanBright,
            hint: _memoryTrend(telemetry.serverHistory),
            sparkline: memorySpots.length >= 2
                ? AetherSparkline(
                    points: memorySpots,
                    color: AetherPalette.cyanBright,
                  )
                : null,
          ),
          KpiCard(
            label: 'Runtime goroutines',
            value: '${server.goroutines}',
            unit: 'go',
            icon: Icons.hub,
            accent: AetherPalette.emeraldBright,
            hint: 'live goroutines in the master',
          ),
          KpiCard(
            label: 'Server uptime',
            value: _formatUptime(server.uptimeSeconds.toInt()),
            unit: 'up',
            icon: Icons.timer_outlined,
            accent: AetherPalette.amber,
            hint: 'since the master started',
          ),
        ],
      ),
    );
  }

  /// Per-agent CPU and memory. This is the reading that matters most when a run
  /// fans out across a larger fleet.
  Widget _buildAgentResources() {
    final agents = widget.telemetry.agents;

    return AetherPanel(
      title: 'Agent resources',
      accent: AetherPalette.emeraldBright,
      padding: EdgeInsets.zero,
      trailing: TelemetryText(
        '${agents.length} online',
        size: 11,
        color: AetherPalette.textMuted,
      ),
      child: agents.isEmpty
          ? const ConsoleMessage(
              'No agents are connected',
              icon: Icons.dns_outlined,
            )
          : AetherDataTable(
              columns: const [
                DataColumn(label: Text('Agent')),
                DataColumn(label: Text('Host')),
                DataColumn(label: Text('CPU')),
                DataColumn(label: Text('Memory')),
                DataColumn(label: Text('Last report')),
              ],
              rows: [
                for (final agent in agents)
                  DataRow(
                    color: AetherDataTable.rowHighlight,
                    cells: [
                      DataCell(
                        TelemetryText(
                          agent.agentId,
                          size: 12.5,
                          color: AetherPalette.cyanBright,
                        ),
                      ),
                      DataCell(
                        TelemetryText(
                          agent.host.isEmpty ? '—' : agent.host,
                          size: 12.5,
                        ),
                      ),
                      DataCell(_cpuCell(agent.cpuPercent)),
                      DataCell(
                        TelemetryText(_formatBytes(agent.memoryBytes.toInt())),
                      ),
                      DataCell(_lastReportCell(agent.lastReportMs.toInt())),
                    ],
                  ),
              ],
            ),
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
            color: _loadAccent(cpuPercent),
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

  Widget _lastReportCell(int lastReportMs) {
    if (lastReportMs <= 0) {
      return TelemetryText('—', size: 12.5, color: AetherPalette.textMuted);
    }

    final age = DateTime.now().difference(
      DateTime.fromMillisecondsSinceEpoch(lastReportMs),
    );
    final label = age.inSeconds < 60
        ? '${age.inSeconds}s ago'
        : '${age.inMinutes}m ago';

    return TelemetryText(label, size: 12.5, color: AetherPalette.textMuted);
  }

  /// Maps a rolling server history onto evenly spaced sparkline points.
  List<FlSpot> _serverSpots(
    List<ServerResourceSample> samples,
    double Function(ServerResourceSample sample) select,
  ) {
    return [
      for (var i = 0; i < samples.length; i++)
        FlSpot(i.toDouble(), select(samples[i])),
    ];
  }

  Widget _buildRegistryBreakdown(List<Agent> agents, List<Target> targets) {
    return AetherPanel(
      title: 'Registry breakdown',
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: _BreakdownColumn(
              title: 'Agent fleet status',
              entries: _agentEntries(agents),
              empty: 'No agents are connected',
            ),
          ),
          const SizedBox(width: 32),
          Expanded(
            child: _BreakdownColumn(
              title: 'Target registry mix',
              entries: _targetEntries(targets),
              empty: 'No targets registered',
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLiveDispatch(List<RunningTestExecution> executions) {
    return AetherPanel(
      title: 'Live dispatch',
      accent: AetherPalette.emeraldBright,
      padding: EdgeInsets.zero,
      child: executions.isEmpty
          ? const ConsoleMessage(
              'No tests are running',
              icon: Icons.play_circle_outline,
            )
          : AetherDataTable(
              columns: const [
                DataColumn(label: Text('Execution')),
                DataColumn(label: Text('Use case')),
                DataColumn(label: Text('Target')),
                DataColumn(label: Text('Agents')),
                DataColumn(label: Text('Users/agent')),
                DataColumn(label: Text('Started')),
              ],
              rows: [
                for (final execution in executions)
                  DataRow(
                    color: AetherDataTable.rowHighlight,
                    cells: [
                      DataCell(
                        TelemetryText(
                          execution.testExecutionId,
                          size: 12.5,
                          color: AetherPalette.cyanBright,
                        ),
                      ),
                      DataCell(
                        Text(
                          execution.useCaseName.isEmpty
                              ? execution.useCaseId
                              : execution.useCaseName,
                        ),
                      ),
                      DataCell(
                        Text(
                          execution.targetName.isEmpty
                              ? execution.targetId
                              : execution.targetName,
                        ),
                      ),
                      DataCell(TelemetryText('${execution.numberOfAgents}')),
                      DataCell(
                        TelemetryText('${execution.simulatedUsersPerAgent}'),
                      ),
                      DataCell(
                        TelemetryText(
                          DateTime.fromMillisecondsSinceEpoch(
                            execution.startTime.toInt() * 1000,
                          ).toLocal().toString(),
                          size: 12.5,
                          color: AetherPalette.textMuted,
                        ),
                      ),
                    ],
                  ),
              ],
            ),
    );
  }
}

/// Vertical list of labelled meters for a single registry dimension.
class _BreakdownColumn extends StatelessWidget {
  const _BreakdownColumn({
    required this.title,
    required this.entries,
    required this.empty,
  });

  final String title;
  final List<_BreakdownEntry> entries;
  final String empty;

  @override
  Widget build(BuildContext context) {
    final total = entries.fold<int>(0, (sum, entry) => sum + entry.count);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title.toUpperCase(), style: AetherText.sectionLabel),
        const SizedBox(height: 14),
        if (entries.isEmpty)
          Text(
            empty,
            style: const TextStyle(
              color: AetherPalette.textMuted,
              fontSize: 12.5,
            ),
          )
        else
          for (final entry in entries)
            Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: _BreakdownRow(entry: entry, total: total),
            ),
      ],
    );
  }
}

class _BreakdownRow extends StatelessWidget {
  const _BreakdownRow({required this.entry, required this.total});

  final _BreakdownEntry entry;
  final int total;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          flex: 3,
          child: Text(
            entry.label.toUpperCase(),
            overflow: TextOverflow.ellipsis,
            style: AetherText.mono(size: 11.5),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          flex: 4,
          child: MeterBar(
            fraction: total == 0 ? 0 : entry.count / total,
            color: entry.color,
          ),
        ),
        const SizedBox(width: 12),
        SizedBox(
          width: 32,
          child: TelemetryText(
            '${entry.count}',
            align: TextAlign.right,
            weight: FontWeight.w700,
            color: entry.color,
          ),
        ),
      ],
    );
  }
}

class _BreakdownEntry {
  const _BreakdownEntry({
    required this.label,
    required this.count,
    required this.color,
  });

  final String label;
  final int count;
  final Color color;
}

/// Explains where the (future) telemetry for this console will come from.
class _TelemetryPerspectiveNote extends StatelessWidget {
  const _TelemetryPerspectiveNote();

  static const _lines = [
    'Control plane and agent fleet are colocated; targets sit outside the '
        'agent cluster boundary.',
    'Latency, throughput and error values are captured from the outbound '
        'synthetic agent perspective.',
    'Server-side target metrics are intentionally out of scope — none are '
        'scraped by the control plane.',
  ];

  @override
  Widget build(BuildContext context) {
    return AetherPanel(
      title: 'Telemetry perspective',
      accent: AetherPalette.amber,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (final line in _lines)
            Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(top: 6),
                    child: StatusDot(color: AetherPalette.amber, size: 6),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      line,
                      style: const TextStyle(
                        color: AetherPalette.textMuted,
                        fontSize: 12.5,
                        height: 1.5,
                      ),
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

/// Semantic accent for a CPU load percentage.
Color _loadAccent(double cpuPercent) {
  if (cpuPercent >= 85) {
    return AetherPalette.crimson;
  }
  if (cpuPercent >= 65) {
    return AetherPalette.amber;
  }
  return AetherPalette.emeraldBright;
}

/// Labels the retained sparkline window as a bare duration, for example `5m`.
String _windowLabel(int samples) {
  final seconds = samples * _serverPollInterval.inSeconds;

  if (seconds < 60) {
    return '${seconds}s';
  }
  if (seconds < 3600) {
    return '${seconds ~/ 60}m';
  }
  return '${(seconds / 3600).toStringAsFixed(1)}h';
}

/// Describes how the master memory footprint moved across the retained window,
/// so growth is legible without reading the sparkline.
String _memoryTrend(List<ServerResourceSample> history) {
  if (history.length < 2) {
    return 'resident set · awaiting samples';
  }

  final delta = history.last.memoryBytes - history.first.memoryBytes;
  final sign = delta < 0 ? '-' : '+';

  return '$sign${_formatBytes(delta.abs())} over ${_windowLabel(history.length)}';
}

/// Human-readable byte size using binary units.
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

/// Compact uptime label for the control plane.
String _formatUptime(int seconds) {
  if (seconds <= 0) {
    return '0s';
  }

  final days = seconds ~/ 86400;
  final hours = (seconds % 86400) ~/ 3600;
  final minutes = (seconds % 3600) ~/ 60;
  final secs = seconds % 60;

  if (days > 0) {
    return '${days}d ${hours}h';
  }
  if (hours > 0) {
    return '${hours}h ${minutes}m';
  }
  if (minutes > 0) {
    return '${minutes}m ${secs}s';
  }
  return '${secs}s';
}

String _protocolSplit(List<Target> targets) {
  if (targets.isEmpty) {
    return 'no targets registered';
  }

  final counts = <String, int>{};
  for (final target in targets) {
    final key = target.protocol.trim().isEmpty
        ? 'unknown'
        : target.protocol.trim().toLowerCase();
    counts[key] = (counts[key] ?? 0) + 1;
  }

  return counts.entries
      .map((entry) => '${entry.value} ${entry.key}')
      .join(' · ');
}

List<_BreakdownEntry> _agentEntries(List<Agent> agents) {
  final counts = <String, int>{};
  for (final agent in agents) {
    final key = agent.status.trim().isEmpty
        ? 'unknown'
        : agent.status.trim().toLowerCase();
    counts[key] = (counts[key] ?? 0) + 1;
  }

  return [
    for (final entry in counts.entries)
      _BreakdownEntry(
        label: entry.key,
        count: entry.value,
        color: statusColor(entry.key),
      ),
  ];
}

List<_BreakdownEntry> _targetEntries(List<Target> targets) {
  final counts = <String, int>{};
  for (final target in targets) {
    final key = target.protocol.trim().isEmpty
        ? 'unknown'
        : target.protocol.trim().toLowerCase();
    counts[key] = (counts[key] ?? 0) + 1;
  }

  return [
    for (final entry in counts.entries)
      _BreakdownEntry(
        label: entry.key,
        count: entry.value,
        color: protocolColor(entry.key),
      ),
  ];
}
