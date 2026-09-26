import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:grpc/grpc.dart';

import '../generated/generated.dart';
import 'service.dart';

/// Loads and mutates the agents registered with the master.
class AgentsRepository extends ChangeNotifier {
  final Backend backend;

  AgentsRepository({required this.backend});

  List<Agent> agents = const [];
  bool loading = false;
  String? error;

  Future<void> load() async {
    loading = true;
    error = null;
    notifyListeners();

    try {
      final response = await backend.master.getAllAgents(GetAllAgentsRequest());
      agents = List<Agent>.from(response.agents);
    } on GrpcError catch (e) {
      backend.handleError(e);
      error = e.message ?? 'Failed to load agents';
    } catch (e) {
      error = 'Failed to load agents: $e';
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<String?> update(Agent agent) async {
    try {
      final response = await backend.master.updateAgent(UpdateAgentRequest(agent: agent));
      if (!response.success) {
        return 'Failed to update agent';
      }
      await load();
      return null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      return e.message ?? 'Failed to update agent';
    }
  }

  Future<String?> delete(Agent agent) async {
    try {
      final response = await backend.master.deleteAgent(DeleteAgentRequest(agent: agent));
      if (!response.success) {
        return 'Failed to delete agent';
      }
      await load();
      return null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      return e.message ?? 'Failed to delete agent';
    }
  }

  Future<String?> clearAll() async {
    try {
      final response = await backend.master.clearAllAgents(ClearAllAgentsRequest());
      if (!response.success) {
        return response.message.isEmpty ? 'Failed to clear agents' : response.message;
      }
      await load();
      return null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      return e.message ?? 'Failed to clear agents';
    }
  }

  Timer? _timer;

  /// Refreshes the agent list every [interval] until [stopPolling] is called,
  /// so agents that connect or disconnect show up without a manual refresh.
  void startPolling({Duration interval = const Duration(seconds: 3)}) {
    _timer ??= Timer.periodic(interval, (_) => load());
  }

  void stopPolling() {
    _timer?.cancel();
    _timer = null;
  }
}

/// Loads and mutates the test targets.
class TargetsRepository extends ChangeNotifier {
  final Backend backend;

  TargetsRepository({required this.backend});

  List<Target> targets = const [];
  bool loading = false;
  String? error;

  Future<void> load() async {
    loading = true;
    error = null;
    notifyListeners();

    try {
      final response = await backend.master.getAllTargets(GetAllTargetsRequest());
      targets = List<Target>.from(response.targets);
    } on GrpcError catch (e) {
      backend.handleError(e);
      error = e.message ?? 'Failed to load targets';
    } catch (e) {
      error = 'Failed to load targets: $e';
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<String?> save(Target target, {required bool isNew}) async {
    try {
      if (isNew) {
        final response = await backend.master.registerTarget(RegisterTargetRequest(target: target));
        if (!response.success) {
          return response.message.isEmpty ? 'Failed to register target' : response.message;
        }
      } else {
        final response = await backend.master.updateTarget(UpdateTargetRequest(target: target));
        if (!response.success) {
          return response.message.isEmpty ? 'Failed to update target' : response.message;
        }
      }
      await load();
      return null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      return e.message ?? 'Failed to save target';
    }
  }

  Future<String?> delete(String id) async {
    try {
      final response = await backend.master.deleteTarget(DeleteTargetRequest(id: id));
      if (!response.success) {
        return response.message.isEmpty ? 'Failed to delete target' : response.message;
      }
      await load();
      return null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      return e.message ?? 'Failed to delete target';
    }
  }
}

/// Loads the (static) list of use cases exposed by the master.
class UseCasesRepository extends ChangeNotifier {
  final Backend backend;

  UseCasesRepository({required this.backend});

  List<UseCase> useCases = const [];
  bool loading = false;
  String? error;

  Future<void> load() async {
    loading = true;
    error = null;
    notifyListeners();

    try {
      final response = await backend.master.getUseCases(GetUseCasesRequest());
      useCases = List<UseCase>.from(response.useCases);
    } on GrpcError catch (e) {
      backend.handleError(e);
      error = e.message ?? 'Failed to load use cases';
    } catch (e) {
      error = 'Failed to load use cases: $e';
    } finally {
      loading = false;
      notifyListeners();
    }
  }
}

/// Loads the running test executions and can start/stop them.
class ExecutionsRepository extends ChangeNotifier {
  final Backend backend;

  ExecutionsRepository({required this.backend});

  List<RunningTestExecution> executions = const [];
  bool loading = false;
  String? error;
  Timer? _timer;

  Future<void> load() async {
    loading = true;
    notifyListeners();

    try {
      final response = await backend.master.getRunningTests(GetRunningTestsRequest());
      executions = List<RunningTestExecution>.from(response.tests);
      error = null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      error = e.message ?? 'Failed to load running tests';
    } catch (e) {
      error = 'Failed to load running tests: $e';
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<String?> start({
    required String useCaseId,
    required String targetId,
    required int simulatedUsersPerAgent,
    required String agentSelection,
  }) async {
    try {
      final response = await backend.master.startTestExecution(
        StartTestExecutionRequest(
          useCaseId: useCaseId,
          targetId: targetId,
          simulatedUsersPerAgent: simulatedUsersPerAgent,
          agentSelection: agentSelection,
        ),
      );

      if (!response.success) {
        return response.message.isEmpty ? 'Failed to start test' : response.message;
      }

      await load();
      return null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      return e.message ?? 'Failed to start test';
    }
  }

  Future<String?> stop(String testExecutionId) async {
    try {
      final response = await backend.master.stopTestExecution(
        StopTestExecutionRequest(testExecutionId: testExecutionId),
      );

      if (!response.success) {
        return response.message.isEmpty ? 'Failed to stop test' : response.message;
      }

      await load();
      return null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      return e.message ?? 'Failed to stop test';
    }
  }

  /// Refreshes the running tests every [interval] until [stopPolling] is called.
  void startPolling({Duration interval = const Duration(seconds: 3)}) {
    _timer ??= Timer.periodic(interval, (_) => load());
  }

  void stopPolling() {
    _timer?.cancel();
    _timer = null;
  }

  @override
  void dispose() {
    stopPolling();
    super.dispose();
  }
}

/// Loads the registered users.
class UsersRepository extends ChangeNotifier {
  final Backend backend;

  UsersRepository({required this.backend});

  List<User> users = const [];
  bool loading = false;
  String? error;

  Future<void> load() async {
    loading = true;
    error = null;
    notifyListeners();

    try {
      final response = await backend.auth.getAllUsers(GetAllUsersRequest());
      users = List<User>.from(response.users);
    } on GrpcError catch (e) {
      backend.handleError(e);
      error = e.message ?? 'Failed to load users';
    } catch (e) {
      error = 'Failed to load users: $e';
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<String?> delete(String id) async {
    try {
      final response = await backend.auth.deleteUser(DeleteUserRequest(id: id));
      if (!response.success) {
        return response.message.isEmpty ? 'Failed to delete user' : response.message;
      }
      await load();
      return null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      return e.message ?? 'Failed to delete user';
    }
  }
}

/// Loads the agent-perspective telemetry for a single test execution.
///
/// The master aggregates the cumulative reports the agents push over their
/// stream, so this is available without Prometheus or Grafana.
class TestMetricsRepository extends ChangeNotifier {
  final Backend backend;

  TestMetricsRepository({required this.backend});

  GetTestMetricsResponse? response;
  bool loading = false;
  String? error;
  Timer? _timer;

  List<TestMetricsSample> get samples => response?.samples ?? const [];
  TestMetricsSummary? get summary => response?.summary;
  List<TestMetricsWorker> get workers => response?.workers ?? const [];

  Future<void> load(String testExecutionId) async {
    loading = true;
    notifyListeners();

    try {
      final result = await backend.master.getTestMetrics(
        GetTestMetricsRequest(testExecutionId: testExecutionId),
      );

      if (!result.found) {
        response = null;
        error = result.message.isEmpty
            ? 'No metrics available for this test execution'
            : result.message;
      } else {
        response = result;
        error = null;
      }
    } on GrpcError catch (e) {
      backend.handleError(e);
      error = e.message ?? 'Failed to load test metrics';
    } catch (e) {
      error = 'Failed to load test metrics: $e';
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  /// Refreshes the metrics every [interval] until [stopPolling] is called.
  void startPolling(
    String testExecutionId, {
    Duration interval = const Duration(seconds: 3),
  }) {
    stopPolling();
    _timer = Timer.periodic(interval, (_) => load(testExecutionId));
  }

  void stopPolling() {
    _timer?.cancel();
    _timer = null;
  }

  /// One-shot fetch that does not touch the shared drilldown state. Used when
  /// exporting a run straight from the history list.
  Future<GetTestMetricsResponse> fetch(String testExecutionId) async {
    final result = await backend.master.getTestMetrics(
      GetTestMetricsRequest(testExecutionId: testExecutionId),
    );

    if (!result.found) {
      throw StateError(
        result.message.isEmpty
            ? 'No metrics available for this test execution'
            : result.message,
      );
    }

    return result;
  }

  /// Clears the current response so a new drilldown does not render stale data.
  void reset() {
    response = null;
    error = null;
    loading = false;
  }

  @override
  void dispose() {
    stopPolling();
    super.dispose();
  }
}

/// Loads the recently finished test executions for the history panel.
///
/// Finished runs are retained by the master (in memory and on disk), so a run
/// whose live export was missed can still be reopened and exported later.
class HistoryRepository extends ChangeNotifier {
  final Backend backend;

  HistoryRepository({required this.backend});

  List<TestRunSummary> runs = const [];
  bool loading = false;
  String? error;
  Timer? _timer;

  Future<void> load({int limit = 0}) async {
    loading = true;
    notifyListeners();

    try {
      final response = await backend.master.getTestHistory(
        GetTestHistoryRequest(limit: limit),
      );
      runs = List<TestRunSummary>.from(response.runs);
      error = null;
    } on GrpcError catch (e) {
      backend.handleError(e);
      error = e.message ?? 'Failed to load test history';
    } catch (e) {
      error = 'Failed to load test history: $e';
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  /// Refreshes the history every [interval] until [stopPolling] is called.
  void startPolling({
    Duration interval = const Duration(seconds: 5),
    int limit = 0,
  }) {
    _timer ??= Timer.periodic(interval, (_) => load(limit: limit));
  }

  void stopPolling() {
    _timer?.cancel();
    _timer = null;
  }

  @override
  void dispose() {
    stopPolling();
    super.dispose();
  }
}
