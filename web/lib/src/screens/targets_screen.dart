import 'package:flutter/material.dart';

import '../data/repo.dart';
import '../generated/generated.dart';
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
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
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
      appBar: AppBar(
        title: const Text('Targets'),
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
            return const Center(child: CircularProgressIndicator());
          }

          if (repository.error != null && repository.targets.isEmpty) {
            return Center(child: Text(repository.error!));
          }

          if (repository.targets.isEmpty) {
            return const Center(child: Text('No targets registered'));
          }

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: DataTable(
              columns: const [
                DataColumn(label: Text('Name')),
                DataColumn(label: Text('Address')),
                DataColumn(label: Text('Protocol')),
                DataColumn(label: Text('Connection reuse')),
                DataColumn(label: Text('SNI')),
                DataColumn(label: Text('Actions')),
              ],
              rows: repository.targets
                  .map(
                    (target) => DataRow(
                      cells: [
                        DataCell(Text(target.name)),
                        DataCell(Text(target.address)),
                        DataCell(Text(target.protocol)),
                        DataCell(
                          Icon(
                            target.connectionReuseEnabled
                                ? Icons.check_circle
                                : Icons.cancel_outlined,
                          ),
                        ),
                        DataCell(Text(target.sni)),
                        DataCell(
                          Row(
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
                  )
                  .toList(),
            ),
          );
        },
      ),
    );
  }
}
