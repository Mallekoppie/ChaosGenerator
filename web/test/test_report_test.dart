import 'package:chaos_master_web/src/data/test_report.dart';
import 'package:chaos_master_web/src/generated/generated.dart';
import 'package:fixnum/fixnum.dart';
import 'package:test/test.dart';

void main() {
  GetTestMetricsResponse buildResponse() => GetTestMetricsResponse(
        found: true,
        summary: TestMetricsSummary(
          testExecutionId: 'test-123',
          useCaseId: 'uc7',
          useCaseName: 'Report Generation',
          targetId: 'target-1',
          targetName: 'Payments',
          targetAddress: 'https://payments.internal:8443',
          targetProtocol: 'https',
          simulatedUsersPerAgent: 10,
          numberOfAgents: 2,
          startTime: Int64(1700000000),
          endTime: Int64(1700000060),
          durationMs: Int64(60000),
          running: false,
          totalRequests: Int64(150),
          totalErrors: Int64(5),
          errorRate: 5 / 150,
          successRequests: Int64(145),
          clientErrors: Int64(1),
          serverErrors: Int64(3),
          timeouts: Int64(1),
          connectionResets: Int64(0),
          connectionErrors: Int64(0),
          p50Ms: 12.5,
          p90Ms: 30,
          p95Ms: 40,
          p99Ms: 55,
          bytesIn: Int64(2048),
          bytesOut: Int64(4096),
          errorsByCode: [
            MapEntry('502', Int64(3)),
            MapEntry('504', Int64(1)),
          ],
        ),
        workers: [
          TestMetricsWorker(
            agentId: 'agent-a',
            online: false,
            totalRequests: Int64(100),
            successRequests: Int64(95),
            clientErrors: Int64(1),
            serverErrors: Int64(3),
            timeouts: Int64(1),
            avgLatencyMs: 20,
            observedP99Ms: 50,
            cpuPercent: 10,
            memoryBytes: Int64(1024),
            metricsScrapes: Int64(60),
            bytesIn: Int64(1024),
            bytesOut: Int64(2048),
            errorRate: 0.05,
          ),
          TestMetricsWorker(
            agentId: 'agent-b',
            online: false,
            totalRequests: Int64(50),
            successRequests: Int64(50),
            avgLatencyMs: 5,
            observedP99Ms: 10,
            cpuPercent: 4,
            memoryBytes: Int64(512),
            metricsScrapes: Int64(60),
            bytesIn: Int64(1024),
            bytesOut: Int64(2048),
          ),
        ],
      );

  test('report includes the required documentation sections', () {
    final report = buildTestReportMarkdown(buildResponse());

    expect(report, contains('# Test Run Report — test-123'));
    expect(report, contains('## Run overview'));
    expect(report, contains('Use case executed'));
    expect(report, contains('Report Generation (uc7)'));
    expect(report, contains('Test start time'));
    expect(report, contains('Test end time'));
    expect(report, contains('Test duration'));
    expect(report, contains('Number of agents'));
    expect(report, contains('## Fleet totals'));
    expect(report, contains('Total requests'));
    expect(report, contains('Requests per agent'));
    expect(report, contains('## Failures by code'));
    expect(report, contains('## Per-agent metrics'));
  });

  test('report lists each agent and a totals row that sums them', () {
    final report = buildTestReportMarkdown(buildResponse());

    expect(report, contains('agent-a'));
    expect(report, contains('agent-b'));
    expect(report, contains('Totals (2 agents)'));
    // Fleet total requests, and the per-agent breakdown.
    expect(report, contains('150'));
    expect(report, contains('agent-a: 100'));
    expect(report, contains('agent-b: 50'));
  });

  test('report formats the run duration', () {
    final report = buildTestReportMarkdown(buildResponse());

    // 60,000ms renders as a minute, not a raw millisecond count.
    expect(report, contains('1m 0s'));
  });

  test('report file name is derived from the execution id', () {
    expect(testReportFileName('test-123'), 'test-123-report.md');
  });
}
