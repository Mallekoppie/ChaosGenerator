import 'package:flutter/material.dart';

import '../data/repo.dart';
import '../generated/generated.dart';
import '../theme.dart';
import '../widgets/aether_chrome.dart';
import '../widgets/aether_common.dart';
import '../widgets/aether_kpi_card.dart';
import '../widgets/aether_panel.dart';
import '../widgets/aether_table.dart';
import '../widgets/agent_edit_dialog.dart';

class AgentsScreen extends StatefulWidget {
  final AgentsRepository repository;

  const AgentsScreen({required this.repository, super.key});

  @override
  State<AgentsScreen> createState() => _AgentsScreenState();
}

class _AgentsScreenState extends State<AgentsScreen> {
  @override
  void initState() {
    super.initState();
    widget.repository.load();
  }

  void _showMessage(String message) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _edit(Agent agent) async {
    final updated = await showDialog<Agent>(
      context: context,
      builder: (context) => AgentEditDialog(agent: agent),
    );

    if (updated == null) {
      return;
    }

    final error = await widget.repository.update(updated);
    if (error != null) {
      _showMessage(error);
    }
  }

  Future<void> _delete(Agent agent) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete agent'),
        content: Text('Remove agent ${agent.host}:${agent.port}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirmed != true) {
      return;
    }

    final error = await widget.repository.delete(agent);
    if (error != null) {
      _showMessage(error);
    }
  }

  Future<void> _clearAll() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Clear all agents'),
        content: const Text('This removes every agent from the database.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Clear all'),
          ),
        ],
      ),
    );

    if (confirmed != true) {
      return;
    }

    final error = await widget.repository.clearAll();
    if (error != null) {
      _showMessage(error);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: aetherAppBar(
        context,
        title: 'Agents',
        actions: [
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: widget.repository.load,
          ),
          TextButton.icon(
            onPressed: _clearAll,
            icon: const Icon(Icons.delete_sweep),
            label: const Text('Clear all'),
          ),
        ],
      ),
      body: ListenableBuilder(
        listenable: widget.repository,
        builder: (context, _) {
          final repository = widget.repository;

          if (repository.loading && repository.agents.isEmpty) {
            return const ConsoleLoading();
          }

          if (repository.error != null && repository.agents.isEmpty) {
            return ConsoleMessage(
              repository.error!,
              icon: Icons.error_outline,
              color: AetherPalette.crimson,
            );
          }

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              // NOTE: placeholder telemetry. These four readings come from the
              // PRD and are hard-coded until the master exposes a metrics API.
              const SectionHeading(
                label: 'Fleet telemetry',
                accent: AetherPalette.amber,
                trailing: StatusPill(
                  label: 'Placeholder telemetry',
                  color: AetherPalette.amber,
                ),
              ),
              const SizedBox(height: 12),
              const KpiDeck(
                cards: [
                  KpiCard(
                    label: 'Total synthetic concurrency',
                    value: '420,000',
                    unit: 'req/s',
                    icon: Icons.speed,
                    hint: 'fleet-wide egress estimate',
                  ),
                  KpiCard(
                    label: 'gRPC fleet ping',
                    value: '1.82',
                    unit: 'ms avg',
                    icon: Icons.network_check,
                    accent: AetherPalette.cyanBright,
                    hint: 'median agent to master',
                  ),
                  KpiCard(
                    label: 'Prometheus scrapes',
                    value: '100',
                    unit: '% healthy',
                    icon: Icons.monitor_heart,
                    accent: AetherPalette.emeraldBright,
                    hint: 'scraper targets healthy',
                  ),
                  KpiCard(
                    label: 'Chaos status',
                    value: 'IDLE',
                    icon: Icons.bolt,
                    accent: AetherPalette.amber,
                    hint: 'no drill in progress',
                  ),
                ],
              ),
              const SizedBox(height: 16),
              AetherPanel(
                title: 'Agent fleet',
                padding: EdgeInsets.zero,
                trailing: TelemetryText(
                  '${repository.agents.length} registered',
                  size: 11,
                  color: AetherPalette.textMuted,
                ),
                child: repository.agents.isEmpty
                    ? const ConsoleMessage(
                        'No agents are connected',
                        icon: Icons.dns_outlined,
                      )
                    : AetherDataTable(
                        columns: const [
                          DataColumn(label: Text('ID')),
                          DataColumn(label: Text('Host')),
                          DataColumn(label: Text('Port')),
                          DataColumn(label: Text('Metrics')),
                          DataColumn(label: Text('Enabled')),
                          DataColumn(label: Text('Status')),
                          DataColumn(label: Text('Actions')),
                        ],
                        rows: [
                          for (final agent in repository.agents)
                            DataRow(
                              color: AetherDataTable.rowHighlight,
                              cells: [
                                DataCell(
                                  TelemetryText(
                                    agent.id,
                                    size: 12.5,
                                    color: AetherPalette.cyanBright,
                                  ),
                                ),
                                DataCell(TelemetryText(agent.host, size: 12.5)),
                                DataCell(TelemetryText('${agent.port}')),
                                DataCell(TelemetryText('${agent.metricsPort}')),
                                DataCell(boolCell(agent.enabled)),
                                DataCell(
                                  TelemetryText(
                                    agent.status,
                                    size: 12.5,
                                    weight: FontWeight.w600,
                                    color: statusColor(agent.status),
                                  ),
                                ),
                                DataCell(
                                  Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        tooltip: 'Edit',
                                        icon: const Icon(Icons.edit),
                                        onPressed: () => _edit(agent),
                                      ),
                                      IconButton(
                                        tooltip: 'Delete',
                                        icon: const Icon(Icons.delete),
                                        onPressed: () => _delete(agent),
                                      ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                        ],
                      ),
              ),
            ],
          );
        },
      ),
    );
  }
}
