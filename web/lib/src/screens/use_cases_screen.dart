import 'package:flutter/material.dart';

import '../data/repo.dart';

class UseCasesScreen extends StatefulWidget {
  final UseCasesRepository repository;

  const UseCasesScreen({required this.repository, super.key});

  @override
  State<UseCasesScreen> createState() => _UseCasesScreenState();
}

class _UseCasesScreenState extends State<UseCasesScreen> {
  @override
  void initState() {
    super.initState();
    widget.repository.load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Use Cases'),
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

          if (repository.loading && repository.useCases.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (repository.error != null && repository.useCases.isEmpty) {
            return Center(child: Text(repository.error!));
          }

          return ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: repository.useCases.length,
            itemBuilder: (context, index) {
              final useCase = repository.useCases[index];
              return Card(
                child: ListTile(
                  leading: Chip(label: Text(useCase.method)),
                  title: Text('${useCase.name}  (${useCase.id})'),
                  subtitle: Text('${useCase.path}\n${useCase.description}'),
                  isThreeLine: true,
                ),
              );
            },
          );
        },
      ),
    );
  }
}
