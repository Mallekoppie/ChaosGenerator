import 'package:flutter/material.dart';

import '../data/repo.dart';
import '../generated/generated.dart';
import '../theme.dart';
import '../widgets/aether_chrome.dart';
import '../widgets/aether_common.dart';
import '../widgets/aether_panel.dart';
import '../widgets/aether_table.dart';

class UsersScreen extends StatefulWidget {
  final UsersRepository repository;

  const UsersScreen({required this.repository, super.key});

  @override
  State<UsersScreen> createState() => _UsersScreenState();
}

class _UsersScreenState extends State<UsersScreen> {
  @override
  void initState() {
    super.initState();
    widget.repository.load();
  }

  Future<void> _delete(User user) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete user'),
        content: Text('Remove user ${user.username}?'),
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

    final error = await widget.repository.delete(user.id);
    if (error != null && mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(error)));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: aetherAppBar(
        context,
        title: 'Users',
        actions: [
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: widget.repository.load,
          ),
        ],
      ),
      body: ListenableBuilder(
        listenable: widget.repository,
        builder: (context, _) {
          final repository = widget.repository;

          if (repository.loading && repository.users.isEmpty) {
            return const ConsoleLoading();
          }

          if (repository.error != null && repository.users.isEmpty) {
            return ConsoleMessage(
              repository.error!,
              icon: Icons.error_outline,
              color: AetherPalette.crimson,
            );
          }

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              AetherPanel(
                title: 'Operator directory',
                padding: EdgeInsets.zero,
                trailing: TelemetryText(
                  '${repository.users.length} accounts',
                  size: 11,
                  color: AetherPalette.textMuted,
                ),
                child: repository.users.isEmpty
                    ? const ConsoleMessage(
                        'No users are registered',
                        icon: Icons.people_outline,
                      )
                    : AetherDataTable(
                        columns: const [
                          DataColumn(label: Text('Username')),
                          DataColumn(label: Text('Created')),
                          DataColumn(label: Text('Actions')),
                        ],
                        rows: [
                          for (final user in repository.users)
                            DataRow(
                              color: AetherDataTable.rowHighlight,
                              cells: [
                                DataCell(
                                  Text(
                                    user.username,
                                    style: const TextStyle(
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                ),
                                DataCell(
                                  TelemetryText(
                                    DateTime.fromMillisecondsSinceEpoch(
                                      user.createdAt.toInt() * 1000,
                                    ).toLocal().toString(),
                                    size: 12.5,
                                    color: AetherPalette.textMuted,
                                  ),
                                ),
                                DataCell(
                                  IconButton(
                                    tooltip: 'Delete',
                                    icon: const Icon(Icons.delete),
                                    onPressed: () => _delete(user),
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
