import 'package:flutter/material.dart';

import '../data/repo.dart';

class TestsScreen extends StatefulWidget {
  final ExecutionsRepository executions;
  final TargetsRepository targets;
  final UseCasesRepository useCases;
  final AgentsRepository agents;

  const TestsScreen({
    required this.executions,
    required this.targets,
    required this.useCases,
    required this.agents,
    super.key,
  });

  @override
  State<TestsScreen> createState() => _TestsScreenState();
}

class _TestsScreenState extends State<TestsScreen> {
  final _usersController = TextEditingController(text: '1');
  final _agentsController = TextEditingController(text: 'all');

  String _selectedUseCaseId = '';
  String _selectedTargetId = '';
  bool _starting = false;

  @override
  void initState() {
    super.initState();

    widget.useCases.load();
    widget.targets.load();
    widget.agents.load();
    widget.executions.load();
    widget.executions.startPolling();
  }

  @override
  void dispose() {
    widget.executions.stopPolling();
    _usersController.dispose();
    _agentsController.dispose();
    super.dispose();
  }

  void _showMessage(String message) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _start() async {
    final users = int.tryParse(_usersController.text) ?? 0;

    if (_selectedUseCaseId.isEmpty ||
        _selectedTargetId.isEmpty ||
        users <= 0) {
      _showMessage('Select a use case, a target and a positive number of users');
      return;
    }

    setState(() => _starting = true);

    final selection = _agentsController.text.trim();
    final error = await widget.executions.start(
      useCaseId: _selectedUseCaseId,
      targetId: _selectedTargetId,
      simulatedUsersPerAgent: users,
      agentSelection: selection.isEmpty ? 'all' : selection,
    );

    if (!mounted) {
      return;
    }

    setState(() => _starting = false);
    _showMessage(error ?? 'Test started');
  }

  Future<void> _stop(String testExecutionId) async {
    final error = await widget.executions.stop(testExecutionId);
    if (error != null) {
      _showMessage(error);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Tests'),
        actions: [
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: widget.executions.load,
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ListenableBuilder(
              listenable: Listenable.merge(
                [widget.useCases, widget.targets, widget.agents],
              ),
              builder: (context, _) => _buildForm(),
            ),
            const SizedBox(height: 24),
            Text(
              'Running tests',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Expanded(
              child: ListenableBuilder(
                listenable: widget.executions,
                builder: (context, _) => _buildRunning(),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildForm() {
    final useCases = widget.useCases.useCases;
    final targets = widget.targets.targets;

    return Wrap(
      spacing: 16,
      runSpacing: 16,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        SizedBox(
          width: 260,
          child: DropdownButtonFormField<String>(
            initialValue: _selectedUseCaseId.isEmpty ? null : _selectedUseCaseId,
            decoration: const InputDecoration(labelText: 'Use case'),
            items: useCases
                .map(
                  (useCase) => DropdownMenuItem(
                    value: useCase.id,
                    child: Text('${useCase.name} (${useCase.id})'),
                  ),
                )
                .toList(),
            onChanged: (value) =>
                setState(() => _selectedUseCaseId = value ?? ''),
          ),
        ),
        SizedBox(
          width: 240,
          child: DropdownButtonFormField<String>(
            initialValue: _selectedTargetId.isEmpty ? null : _selectedTargetId,
            decoration: const InputDecoration(labelText: 'Target'),
            items: targets
                .map(
                  (target) => DropdownMenuItem(
                    value: target.id,
                    child: Text(target.name),
                  ),
                )
                .toList(),
            onChanged: (value) =>
                setState(() => _selectedTargetId = value ?? ''),
          ),
        ),
        SizedBox(
          width: 160,
          child: TextField(
            controller: _usersController,
            decoration: const InputDecoration(labelText: 'Users per agent'),
            keyboardType: TextInputType.number,
          ),
        ),
        SizedBox(
          width: 240,
          child: TextField(
            controller: _agentsController,
            decoration: const InputDecoration(
              labelText: 'Agents',
              hintText: 'all or comma separated ids',
            ),
          ),
        ),
        FilledButton.icon(
          onPressed: _starting ? null : _start,
          icon: const Icon(Icons.play_arrow),
          label: Text(_starting ? 'Starting...' : 'Start test'),
        ),
      ],
    );
  }

  Widget _buildRunning() {
    final repository = widget.executions;

    if (repository.executions.isEmpty) {
      return const Center(child: Text('No tests are running'));
    }

    return SingleChildScrollView(
      child: DataTable(
        columns: const [
          DataColumn(label: Text('Execution')),
          DataColumn(label: Text('Use case')),
          DataColumn(label: Text('Target')),
          DataColumn(label: Text('Agents')),
          DataColumn(label: Text('Users/agent')),
          DataColumn(label: Text('Started')),
          DataColumn(label: Text('')),
        ],
        rows: repository.executions
            .map(
              (execution) => DataRow(
                cells: [
                  DataCell(Text(execution.testExecutionId)),
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
                  DataCell(Text('${execution.numberOfAgents}')),
                  DataCell(Text('${execution.simulatedUsersPerAgent}')),
                  DataCell(
                    Text(
                      DateTime.fromMillisecondsSinceEpoch(
                        execution.startTime.toInt() * 1000,
                      ).toLocal().toString(),
                    ),
                  ),
                  DataCell(
                    TextButton(
                      onPressed: () => _stop(execution.testExecutionId),
                      child: const Text('Stop'),
                    ),
                  ),
                ],
              ),
            )
            .toList(),
      ),
    );
  }
}
