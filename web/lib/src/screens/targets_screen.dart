import 'package:flutter/material.dart';

import '../data/repo.dart';
import '../generated/generated.dart';
import '../theme.dart';
import '../widgets/aether_chrome.dart';
import '../widgets/aether_common.dart';
import '../widgets/aether_kpi_card.dart';
import '../widgets/aether_panel.dart';
import '../widgets/aether_table.dart';
import '../widgets/target_edit_dialog.dart';

class TargetsScreen extends StatefulWidget {
  final TargetsRepository repository;

  const TargetsScreen({required this.repository, super.key});

  @override
  State<TargetsScreen> createState() => _TargetsScreenState();
}

class _TargetsScreenState extends State<TargetsScreen> {
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

  Future<void> _add() async {
    final target = await showDialog<Target>(
      context: context,
      builder: (context) => const TargetEditDialog(),
    );

    if (target == null) {
      return;
    }

    final error = await widget.repository.save(target, isNew: true);
    if (error != null) {
      _showMessage(error);
    }
  }

  Future<void> _edit(Target target) async {
    final updated = await showDialog<Target>(
      context: context,
      builder: (context) => TargetEditDialog(target: target),
    );

    if (updated == null) {
      return;
    }

    final error = await widget.repository.save(updated, isNew: false);
    if (error != null) {
      _showMessage(error);
    }
  }

  Future<void> _delete(Target target) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete target'),
        content: Text('Remove target ${target.name}?'),
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

    final error = await widget.repository.delete(target.id);
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
        title: 'Targets',
        actions: [
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: widget.repository.load,
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _add,
        tooltip: 'Add target',
        child: const Icon(Icons.add),
      ),
      body: ListenableBuilder(
        listenable: widget.repository,
        builder: (context, _) {
          final repository = widget.repository;

          if (repository.loading && repository.targets.isEmpty) {
            return const ConsoleLoading();
          }

          if (repository.error != null && repository.targets.isEmpty) {
            return ConsoleMessage(
              repository.error!,
              icon: Icons.error_outline,
              color: AetherPalette.crimson,
            );
          }

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              // NOTE: placeholder telemetry. These summary cards come from the
              // PRD and are hard-coded until the master exposes a metrics API.
              const SectionHeading(
                label: 'Registry telemetry',
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
                    label: 'Health status',
                    value: '100',
                    unit: '% OK',
                    icon: Icons.verified_user,
                    accent: AetherPalette.emeraldBright,
                    hint: 'all targets reachable',
                  ),
                  KpiCard(
                    label: 'Protocol split',
                    value: '3',
                    unit: 'protocols',
                    icon: Icons.hub,
                    accent: AetherPalette.cyanBright,
                    hint: 'https · http · grpc',
                  ),
                  KpiCard(
                    label: 'Connection reuse',
                    value: '4/5',
                    unit: 'active',
                    icon: Icons.sync_alt,
                    hint: 'keep-alive enabled',
                  ),
                  KpiCard(
                    label: 'SNI strictness',
                    value: 'ENFORCED',
                    icon: Icons.shield,
                    accent: AetherPalette.amber,
                    hint: 'sni validated targets',
                  ),
                ],
              ),
              const SizedBox(height: 16),
              AetherPanel(
                title: 'Target registry',
                padding: EdgeInsets.zero,
                trailing: TelemetryText(
                  '${repository.targets.length} registered',
                  size: 11,
                  color: AetherPalette.textMuted,
                ),
                child: repository.targets.isEmpty
                    ? const ConsoleMessage(
                        'No targets registered',
                        icon: Icons.my_location_outlined,
                      )
                    : AetherDataTable(
                        columns: const [
                          DataColumn(label: Text('Name')),
                          DataColumn(label: Text('Address')),
                          DataColumn(label: Text('Protocol')),
                          DataColumn(label: Text('Connection reuse')),
                          DataColumn(label: Text('SNI')),
                          DataColumn(label: Text('Actions')),
                        ],
                        rows: [
                          for (final target in repository.targets)
                            DataRow(
                              color: AetherDataTable.rowHighlight,
                              cells: [
                                DataCell(
                                  Text(
                                    target.name,
                                    style: const TextStyle(
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                ),
                                DataCell(
                                  TelemetryText(
                                    target.address,
                                    size: 12.5,
                                    color: AetherPalette.cyanBright,
                                  ),
                                ),
                                DataCell(_ProtocolBadge(target.protocol)),
                                DataCell(
                                  boolCell(target.connectionReuseEnabled),
                                ),
                                DataCell(
                                  TelemetryText(
                                    target.sni,
                                    size: 12.5,
                                    color: AetherPalette.textMuted,
                                  ),
                                ),
                                DataCell(
                                  Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        tooltip: 'Edit',
                                        icon: const Icon(Icons.edit),
                                        onPressed: () => _edit(target),
                                      ),
                                      IconButton(
                                        tooltip: 'Delete',
                                        icon: const Icon(Icons.delete),
                                        onPressed: () => _delete(target),
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

/// Small colour-coded transport badge for a target's protocol.
class _ProtocolBadge extends StatelessWidget {
  const _ProtocolBadge(this.protocol);

  final String protocol;

  @override
  Widget build(BuildContext context) {
    final color = protocolColor(protocol);

    return Align(
      alignment: Alignment.centerLeft,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 4),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.12),
          borderRadius: BorderRadius.circular(7),
          border: Border.all(color: color.withValues(alpha: 0.45)),
        ),
        child: Text(
          protocol.toUpperCase(),
          style: AetherText.mono(
            size: 10.5,
            weight: FontWeight.w700,
            color: color,
            letterSpacing: 1.0,
          ),
        ),
      ),
    );
  }
}
