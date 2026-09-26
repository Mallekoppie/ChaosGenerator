import 'package:chaos_master_web/src/generated/generated.dart';
import 'package:chaos_master_web/src/theme.dart';
import 'package:chaos_master_web/src/widgets/agent_edit_dialog.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

Agent _agent() => Agent()
  ..id = 'agent-1'
  ..host = '10.0.0.5'
  ..port = 9003
  ..metricsPort = 9103
  ..status = 'online'
  ..enabled = true;

/// Opens the dialog from a route so the value popped by Save can be inspected.
Future<Agent> _openAndSave(
  WidgetTester tester,
  Agent agent, {
  String? newHost,
}) async {
  Agent? result;

  await tester.pumpWidget(
    MaterialApp(
      theme: aetherTheme(),
      home: Builder(
        builder: (context) => Scaffold(
          body: Center(
            child: TextButton(
              onPressed: () async {
                result = await showDialog<Agent>(
                  context: context,
                  builder: (_) => AgentEditDialog(agent: agent),
                );
              },
              child: const Text('open'),
            ),
          ),
        ),
      ),
    ),
  );

  await tester.tap(find.text('open'));
  await tester.pumpAndSettle();

  if (newHost != null) {
    final hostField = tester
        .widgetList<TextField>(find.byType(TextField))
        .firstWhere((field) => field.controller?.text == agent.host);
    await tester.enterText(find.byWidget(hostField), newHost);
    await tester.pumpAndSettle();
  }

  await tester.tap(find.text('Save'));
  await tester.pumpAndSettle();

  final saved = result;
  if (saved == null) {
    throw StateError('AgentEditDialog did not return an agent on save');
  }

  return saved;
}

void main() {
  testWidgets('agent status is displayed but cannot be edited', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: aetherTheme(),
        home: Scaffold(body: AgentEditDialog(agent: _agent())),
      ),
    );

    final fields = tester
        .widgetList<TextField>(find.byType(TextField))
        .toList();

    final statusField = fields.singleWhere(
      (field) => field.controller?.text == 'online',
    );
    expect(
      statusField.enabled,
      isFalse,
      reason: 'status is reported by the agent and must not be editable',
    );

    final hostField = fields.singleWhere(
      (field) => field.controller?.text == '10.0.0.5',
    );
    expect(hostField.enabled, isNot(false), reason: 'host must stay editable');
  });

  testWidgets('saving preserves the agent-reported status', (tester) async {
    final saved = await _openAndSave(tester, _agent(), newHost: '10.0.0.9');

    expect(saved.host, '10.0.0.9', reason: 'editable fields must still apply');
    expect(saved.status, 'online', reason: 'status must round-trip unchanged');
  });
}
