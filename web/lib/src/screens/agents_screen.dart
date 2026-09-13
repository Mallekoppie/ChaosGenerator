import 'package:flutter/material.dart';

import '../data/repo.dart';
import '../generated/generated.dart';
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
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
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
      appBar: AppBar(
        title: const Text('Agents'),
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
            return const Center(child: CircularProgressIndicator());
          }

          if (repository.error != null && repository.agents.isEmpty) {
            return Center(child: Text(repository.error!));
          }

          if (repository.agents.isEmpty) {
            return const Center(child: Text('No agents are connected'));
          }

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: DataTable(
              columns: const [
                DataColumn(label: Text('ID')),
                DataColumn(label: Text('Host')),
                DataColumn(label: Text('Port')),
                DataColumn(label: Text('Metrics')),
                DataColumn(label: Text('Enabled')),
                DataColumn(label: Text('Status')),
                DataColumn(label: Text('Actions')),
              ],
              rows: repository.agents
                  .map(
                    (agent) => DataRow(
                      cells: [
                        DataCell(Text(agent.id)),
                        DataCell(Text(agent.host)),
                        DataCell(Text('${agent.port}')),
                        DataCell(Text('${agent.metricsPort}')),
                        DataCell(
                          Icon(
                            agent.enabled
                                ? Icons.check_circle
                                : Icons.cancel_outlined,
                          ),
                        ),
                        DataCell(Text(agent.status)),
                        DataCell(
                          Row(
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
                  )
                  .toList(),
            ),
          );
        },
      ),
    );
  }
}
