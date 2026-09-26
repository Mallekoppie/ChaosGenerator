import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../data/download.dart';
import '../data/repo.dart';
import '../data/test_report.dart';
import '../generated/generated.dart';
import '../theme.dart';
import '../widgets/aether_chrome.dart';
import '../widgets/aether_common.dart';
import '../widgets/aether_panel.dart';
import '../widgets/aether_table.dart';

class TestsScreen extends StatefulWidget {
  final ExecutionsRepository executions;
  final HistoryRepository history;
  final TestMetricsRepository metrics;
  final TargetsRepository targets;
  final UseCasesRepository useCases;
  final AgentsRepository agents;

  const TestsScreen({
    required this.executions,
    required this.history,
    required this.metrics,
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
    widget.history.load();
    widget.history.startPolling();
  }

  @override
  void dispose() {
    widget.executions.stopPolling();
    widget.history.stopPolling();
    _usersController.dispose();
    _agentsController.dispose();
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

  Future<void> _start() async {
    final users = int.tryParse(_usersController.text) ?? 0;

    if (_selectedUseCaseId.isEmpty || _selectedTargetId.isEmpty || users <= 0) {
      _showMessage(
        'Select a use case, a target and a positive number of users',
      );
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
    // A stopped run moves from the live table into the history panel.
    await widget.history.load();
  }

  Future<void> _refresh() async {
    await Future.wait([
      widget.executions.load(),
      widget.history.load(),
    ]);
  }

  /// Fetches the full metrics for one history row and downloads the Markdown
  /// documentation report. The run no longer needs to be live to be exported.
  Future<void> _exportRun(TestRunSummary run) async {
    _showMessage('Preparing report for ${run.testExecutionId}...');

    try {
      final response = await widget.metrics.fetch(run.testExecutionId);
      downloadText(
        testReportFileName(run.testExecutionId),
        buildTestReportMarkdown(response),
        mimeType: 'text/markdown;charset=utf-8',
      );
      _showMessage('Report exported');
    } catch (e) {
      _showMessage('Failed to export report: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: aetherAppBar(
        context,
        title: 'Tests',
        actions: [
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: _refresh,
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            ListenableBuilder(
              listenable: Listenable.merge([
                widget.useCases,
                widget.targets,
                widget.agents,
              ]),
              builder: (context, _) => AetherPanel(
                title: 'Start a test',
                accent: AetherPalette.emeraldBright,
                child: _buildForm(),
              ),
            ),
            const SizedBox(height: 16),
            Expanded(
              flex: 3,
              child: ListenableBuilder(
                listenable: widget.executions,
                builder: (context, _) => AetherPanel(
                  title: 'Running tests',
                  fill: true,
                  padding: EdgeInsets.zero,
                  trailing: const StatusPill(
                    label: 'Auto-refresh: 3s',
                    color: AetherPalette.emeraldBright,
                  ),
                  child: _buildRunning(),
                ),
              ),
            ),
            const SizedBox(height: 16),
            Expanded(
              flex: 2,
              child: ListenableBuilder(
                listenable: widget.history,
                builder: (context, _) => AetherPanel(
                  title: 'Recent runs',
                  fill: true,
                  padding: EdgeInsets.zero,
                  trailing: const StatusPill(
                    label: 'Auto-refresh: 5s',
                    color: AetherPalette.emeraldBright,
                  ),
                  child: _buildHistory(),
                ),
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
            initialValue: _selectedUseCaseId.isEmpty
                ? null
                : _selectedUseCaseId,
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
      return const ConsoleMessage(
        'No tests are running',
        icon: Icons.play_circle_outline,
      );
    }

    return SingleChildScrollView(
      child: AetherDataTable(
        columns: const [
          DataColumn(label: Text('Execution')),
          DataColumn(label: Text('Use case')),
          DataColumn(label: Text('Target')),
          DataColumn(label: Text('Agents')),
          DataColumn(label: Text('Users/agent')),
          DataColumn(label: Text('Started')),
          DataColumn(label: Text('')),
        ],
        rows: [
          for (final execution in repository.executions)
            DataRow(
              color: AetherDataTable.rowHighlight,
              cells: [
                DataCell(
                  Tooltip(
                    message: 'Open metrics drilldown',
                    child: InkWell(
                      onTap: () => context.go(
                        '/tests/metrics/${execution.testExecutionId}',
                      ),
                      child: TelemetryText(
                        execution.testExecutionId,
                        size: 12.5,
                        color: AetherPalette.cyanBright,
                      ),
                    ),
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
                DataCell(TelemetryText('${execution.simulatedUsersPerAgent}')),
                DataCell(
                  TelemetryText(
                    DateTime.fromMillisecondsSinceEpoch(
                      execution.startTime.toInt() * 1000,
                    ).toLocal().toString(),
                    size: 12.5,
                    color: AetherPalette.textMuted,
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
        ],
      ),
    );
  }

  Widget _buildHistory() {
    final repository = widget.history;

    if (repository.runs.isEmpty) {
      return ConsoleMessage(
        repository.loading
            ? 'Loading previous runs...'
            : 'No previous test runs',
        icon: Icons.history,
      );
    }

    return SingleChildScrollView(
      child: AetherDataTable(
        columns: const [
          DataColumn(label: Text('Execution')),
          DataColumn(label: Text('Use case')),
          DataColumn(label: Text('Agents')),
          DataColumn(label: Text('Requests')),
          DataColumn(label: Text('Errors')),
          DataColumn(label: Text('Started')),
          DataColumn(label: Text('Ended')),
          DataColumn(label: Text('Duration')),
          DataColumn(label: Text('')),
        ],
        rows: [
          for (final run in repository.runs)
            DataRow(
              cells: [
                DataCell(
                  Tooltip(
                    message: 'Open metrics drilldown',
                    child: InkWell(
                      onTap: () => context.go(
                        '/tests/metrics/${run.testExecutionId}',
                      ),
                      child: TelemetryText(
                        run.testExecutionId,
                        size: 12.5,
                        color: AetherPalette.cyanBright,
                      ),
                    ),
                  ),
                ),
                DataCell(
                  Text(
                    run.useCaseName.isEmpty ? run.useCaseId : run.useCaseName,
                  ),
                ),
                DataCell(TelemetryText('${run.numberOfAgents}')),
                DataCell(TelemetryText('${run.totalRequests.toInt()}')),
                DataCell(TelemetryText('${run.totalErrors.toInt()}')),
                DataCell(
                  TelemetryText(
                    formatLocalTimestamp(run.startTime.toInt()),
                    size: 12.5,
                    color: AetherPalette.textMuted,
                  ),
                ),
                DataCell(
                  TelemetryText(
                    formatLocalTimestamp(run.endTime.toInt()),
                    size: 12.5,
                    color: AetherPalette.textMuted,
                  ),
                ),
                DataCell(
                  TelemetryText(formatDurationMs(run.durationMs.toInt())),
                ),
                DataCell(
                  Tooltip(
                    message: 'Export documentation report',
                    child: IconButton(
                      icon: const Icon(Icons.download, size: 18),
                      onPressed: () => _exportRun(run),
                    ),
                  ),
                ),
              ],
            ),
        ],
      ),
    );
  }
}
