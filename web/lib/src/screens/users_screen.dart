import 'package:flutter/material.dart';

import '../data/repo.dart';
import '../generated/generated.dart';

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
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(error)));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Users'),
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
            return const Center(child: CircularProgressIndicator());
          }

          if (repository.error != null && repository.users.isEmpty) {
            return Center(child: Text(repository.error!));
          }

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: DataTable(
              columns: const [
                DataColumn(label: Text('Username')),
                DataColumn(label: Text('Created')),
                DataColumn(label: Text('Actions')),
              ],
              rows: repository.users
                  .map(
                    (user) => DataRow(
                      cells: [
                        DataCell(Text(user.username)),
                        DataCell(
                          Text(
                            DateTime.fromMillisecondsSinceEpoch(
                              user.createdAt.toInt() * 1000,
                            ).toLocal().toString(),
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
                  )
                  .toList(),
            ),
          );
        },
      ),
    );
  }
}
