import '../generated/generated.dart';

/// Builds a human-readable Markdown report for one test execution.
///
/// The report is intended for documentation: it summarises the run, the
/// fleet-wide totals and the full per-agent breakdown (with a totals row), so a
/// finished run can be exported at any time — including from history after the
/// live drilldown has gone.
String buildTestReportMarkdown(
  GetTestMetricsResponse response, {
  DateTime? generatedAt,
}) {
  final summary = response.summary;
  final generated = (generatedAt ?? DateTime.now()).toLocal();

  final buffer = StringBuffer();

  buffer.writeln('# Test Run Report — ${summary.testExecutionId}');
  buffer.writeln();
  buffer.writeln('_Generated ${_formatDateTime(generated)} (local time)._');
  buffer.writeln();

  _writeRunOverview(buffer, summary);
  _writeFleetTotals(buffer, summary, response.workers);
  _writeFailuresByCode(buffer, summary);
  _writePerAgentMetrics(buffer, response.workers);

  return buffer.toString();
}

/// Suggested file name for the exported report.
String testReportFileName(String executionId) => '$executionId-report.md';

// --- Sections ----------------------------------------------------------------

void _writeRunOverview(StringBuffer buffer, TestMetricsSummary summary) {
  final useCase = summary.useCaseName.isEmpty
      ? summary.useCaseId
      : '${summary.useCaseName} (${summary.useCaseId})';
  final target = summary.targetName.isEmpty
      ? summary.targetId
      : summary.targetName;
  final targetDetail = [
    target,
    if (summary.targetAddress.isNotEmpty) summary.targetAddress,
    if (summary.targetProtocol.isNotEmpty)
      'over ${summary.targetProtocol.toUpperCase()}',
  ].join(' — ');

  buffer.writeln('## Run overview');
  buffer.writeln();
  buffer.writeln('| Field | Value |');
  buffer.writeln('| --- | --- |');
  _row(buffer, 'Execution ID', summary.testExecutionId);
  _row(buffer, 'Use case executed', useCase);
  _row(buffer, 'Target', targetDetail);
  _row(buffer, 'Test start time', formatLocalTimestamp(summary.startTime.toInt()));
  _row(
    buffer,
    'Test end time',
    summary.running
        ? 'Still running'
        : formatLocalTimestamp(summary.endTime.toInt()),
  );
  _row(buffer, 'Test duration', formatDurationMs(summary.durationMs.toInt()));
  _row(buffer, 'Number of agents', '${summary.numberOfAgents}');
  _row(buffer, 'Users per agent', '${summary.simulatedUsersPerAgent}');
  _row(buffer, 'Status', summary.running ? 'Running' : 'Stopped');
  buffer.writeln();
}

void _writeFleetTotals(
  StringBuffer buffer,
  TestMetricsSummary summary,
  List<TestMetricsWorker> workers,
) {
  final totalRequests = summary.totalRequests.toInt();
  final totalErrors = summary.totalErrors.toInt();
  final successRate = totalRequests == 0
      ? 0.0
      : summary.successRequests.toInt() / totalRequests;
  final durationMs = summary.durationMs.toInt();
  final avgRps = durationMs > 0 ? totalRequests / (durationMs / 1000.0) : 0.0;

  buffer.writeln('## Fleet totals');
  buffer.writeln();
  buffer.writeln('| Metric | Value |');
  buffer.writeln('| --- | --- |');
  _row(buffer, 'Total requests', _formatInt(totalRequests));
  _row(buffer, 'Requests per agent', _requestsPerAgent(workers));
  _row(buffer, 'Successful requests', _formatInt(summary.successRequests.toInt()));
  _row(buffer, 'Success rate', _formatPercent(successRate));
  _row(buffer, 'Total errors', _formatInt(totalErrors));
  _row(buffer, 'Error rate', _formatPercent(summary.errorRate));
  _row(buffer, 'Client errors (4xx)', _formatInt(summary.clientErrors.toInt()));
  _row(buffer, 'Server errors (5xx)', _formatInt(summary.serverErrors.toInt()));
  _row(buffer, 'Timeouts', _formatInt(summary.timeouts.toInt()));
  _row(buffer, 'Connection resets', _formatInt(summary.connectionResets.toInt()));
  _row(buffer, 'Connection errors', _formatInt(summary.connectionErrors.toInt()));
  _row(buffer, 'Average request rate', _formatRate(avgRps));
  _row(buffer, 'Latency p50 / p90 / p95 / p99', _latencyRow(summary));
  _row(buffer, 'Data sent', _formatBytes(summary.bytesOut.toInt()));
  _row(buffer, 'Data received', _formatBytes(summary.bytesIn.toInt()));
  buffer.writeln();
}

void _writeFailuresByCode(StringBuffer buffer, TestMetricsSummary summary) {
  final codes = summary.errorsByCode.map((k, v) => MapEntry(k, v.toInt()));
  if (codes.isEmpty) {
    return;
  }

  final sorted = codes.entries.toList()
    ..sort((a, b) => b.value.compareTo(a.value));

  buffer.writeln('## Failures by code');
  buffer.writeln();
  buffer.writeln('| Code | Count |');
  buffer.writeln('| --- | --- |');
  for (final entry in sorted) {
    _row(buffer, entry.key, _formatInt(entry.value));
  }
  buffer.writeln();
}

void _writePerAgentMetrics(
  StringBuffer buffer,
  List<TestMetricsWorker> workers,
) {
  buffer.writeln('## Per-agent metrics');
  buffer.writeln();

  if (workers.isEmpty) {
    buffer.writeln('_No agents reported for this run._');
    buffer.writeln();
    return;
  }

  const headers = [
    'Agent',
    'Online',
    'Users',
    'Requests',
    'Success',
    '4xx',
    '5xx',
    'Timeouts',
    'Resets',
    'Conn err',
    'Other err',
    'Error rate',
    'Avg ms',
    'p50 ms',
    'p90 ms',
    'p95 ms',
    'p99 ms',
    'CPU %',
    'Memory',
    'Scrapes',
    'Bytes sent',
    'Bytes received',
  ];

  buffer.writeln('| ${headers.join(' | ')} |');
  buffer.writeln('| ${List.filled(headers.length, '---').join(' | ')} |');

  for (final worker in workers) {
    buffer.writeln('| ${_workerCells(worker).join(' | ')} |');
  }

  buffer.writeln('| ${_totalsCells(workers).join(' | ')} |');
  buffer.writeln();
}

List<String> _workerCells(TestMetricsWorker worker) => [
      _cell(worker.agentId),
      worker.online ? 'yes' : 'no',
      '${worker.activeUsers}',
      _formatInt(worker.totalRequests.toInt()),
      _formatInt(worker.successRequests.toInt()),
      _formatInt(worker.clientErrors.toInt()),
      _formatInt(worker.serverErrors.toInt()),
      _formatInt(worker.timeouts.toInt()),
      _formatInt(worker.connectionResets.toInt()),
      _formatInt(worker.connectionErrors.toInt()),
      _formatInt(worker.otherErrors.toInt()),
      _formatPercent(worker.errorRate),
      _formatMs(worker.avgLatencyMs),
      _formatMs(worker.p50Ms),
      _formatMs(worker.p90Ms),
      _formatMs(worker.p95Ms),
      _formatMs(worker.observedP99Ms),
      _formatMs(worker.cpuPercent),
      _formatBytes(worker.memoryBytes.toInt()),
      _formatInt(worker.metricsScrapes.toInt()),
      _formatBytes(worker.bytesOut.toInt()),
      _formatBytes(worker.bytesIn.toInt()),
    ];

List<String> _totalsCells(List<TestMetricsWorker> workers) {
  var requests = 0;
  var success = 0;
  var clientErrors = 0;
  var serverErrors = 0;
  var timeouts = 0;
  var resets = 0;
  var connErrors = 0;
  var otherErrors = 0;
  var cpu = 0.0;
  var memory = 0;
  var scrapes = 0;
  var bytesIn = 0;
  var bytesOut = 0;
  var latencyWeighted = 0.0;

  for (final worker in workers) {
    requests += worker.totalRequests.toInt();
    success += worker.successRequests.toInt();
    clientErrors += worker.clientErrors.toInt();
    serverErrors += worker.serverErrors.toInt();
    timeouts += worker.timeouts.toInt();
    resets += worker.connectionResets.toInt();
    connErrors += worker.connectionErrors.toInt();
    otherErrors += worker.otherErrors.toInt();
    cpu += worker.cpuPercent;
    memory += worker.memoryBytes.toInt();
    scrapes += worker.metricsScrapes.toInt();
    bytesIn += worker.bytesIn.toInt();
    bytesOut += worker.bytesOut.toInt();
    latencyWeighted += worker.avgLatencyMs * worker.totalRequests.toInt();
  }

  final totalErrors =
      clientErrors + serverErrors + timeouts + resets + connErrors + otherErrors;
  final errorRate = requests == 0 ? 0.0 : totalErrors / requests;
  final avgLatency = requests == 0 ? 0.0 : latencyWeighted / requests;

  return [
    '**Totals (${workers.length} agents)**',
    '',
    '',
    '**${_formatInt(requests)}**',
    '**${_formatInt(success)}**',
    '**${_formatInt(clientErrors)}**',
    '**${_formatInt(serverErrors)}**',
    '**${_formatInt(timeouts)}**',
    '**${_formatInt(resets)}**',
    '**${_formatInt(connErrors)}**',
    '**${_formatInt(otherErrors)}**',
    '**${_formatPercent(errorRate)}**',
    '**${_formatMs(avgLatency)}**',
    '', '', '', '', '',
    '**${_formatMs(cpu)}**',
    '**${_formatBytes(memory)}**',
    '**${_formatInt(scrapes)}**',
    '**${_formatBytes(bytesOut)}**',
    '**${_formatBytes(bytesIn)}**',
  ];
}

// --- Helpers -----------------------------------------------------------------

void _row(StringBuffer buffer, String label, String value) {
  buffer.writeln('| ${_cell(label)} | ${_cell(value)} |');
}

/// Escapes a Markdown table cell so user-controlled values cannot break it.
String _cell(String value) =>
    value.replaceAll('|', r'\|').replaceAll('\n', ' ').trim();

String _requestsPerAgent(List<TestMetricsWorker> workers) {
  if (workers.isEmpty) {
    return '—';
  }

  final parts = [
    for (final worker in workers)
      '${worker.agentId}: ${_formatInt(worker.totalRequests.toInt())}',
  ];
  return parts.join(' · ');
}

String _latencyRow(TestMetricsSummary summary) {
  if (summary.totalRequests.toInt() == 0) {
    return '—';
  }

  return '${_formatMs(summary.p50Ms)} / ${_formatMs(summary.p90Ms)} / '
      '${_formatMs(summary.p95Ms)} / ${_formatMs(summary.p99Ms)}';
}

String _formatInt(int value) {
  final negative = value < 0;
  final digits = (negative ? -value : value).toString();
  final buffer = StringBuffer();

  for (var i = 0; i < digits.length; i++) {
    if (i > 0 && (digits.length - i) % 3 == 0) {
      buffer.write(',');
    }
    buffer.write(digits[i]);
  }

  return negative ? '-$buffer' : buffer.toString();
}

/// Formats a duration in milliseconds as a short human-readable string, e.g.
/// `2m 5s`, `1h 3m 9s` or `450ms`.
String formatDurationMs(int milliseconds) {
  if (milliseconds <= 0) {
    return '—';
  }

  final duration = Duration(milliseconds: milliseconds);
  final hours = duration.inHours;
  final minutes = duration.inMinutes % 60;
  final seconds = duration.inSeconds % 60;

  if (hours > 0) {
    return '${hours}h ${minutes}m ${seconds}s';
  }
  if (minutes > 0) {
    return '${minutes}m ${seconds}s';
  }
  if (seconds > 0) {
    final tenths = (duration.inMilliseconds % 1000) ~/ 100;
    return '$seconds.${tenths}s';
  }

  return '${duration.inMilliseconds}ms';
}

/// Formats a Unix timestamp (seconds) as a local date-time, or `—` when unset.
String formatLocalTimestamp(int unixSeconds) {
  if (unixSeconds <= 0) {
    return '—';
  }

  return _formatDateTime(
    DateTime.fromMillisecondsSinceEpoch(unixSeconds * 1000).toLocal(),
  );
}

String _formatDateTime(DateTime dt) {
  String two(int value) => value.toString().padLeft(2, '0');

  return '${dt.year.toString().padLeft(4, '0')}-${two(dt.month)}-${two(dt.day)} '
      '${two(dt.hour)}:${two(dt.minute)}:${two(dt.second)}';
}

String _formatMs(double value) {
  if (value <= 0) {
    return '—';
  }

  return value >= 100 ? value.toStringAsFixed(0) : value.toStringAsFixed(2);
}

String _formatPercent(double value) {
  if (value <= 0) {
    return '0%';
  }

  return '${(value * 100).toStringAsFixed(2)}%';
}

String _formatRate(double value) => '${value.toStringAsFixed(1)} req/s';

String _formatBytes(int bytes) {
  if (bytes <= 0) {
    return '0 B';
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  var value = bytes.toDouble();
  var unit = 0;

  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }

  final rendered = unit == 0
      ? value.toStringAsFixed(0)
      : value.toStringAsFixed(2);
  return '$rendered ${units[unit]}';
}
