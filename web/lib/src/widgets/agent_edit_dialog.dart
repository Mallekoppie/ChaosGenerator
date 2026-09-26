import 'package:flutter/material.dart';

import '../generated/generated.dart';

/// Dialog used to edit an agent's connection details.
class AgentEditDialog extends StatefulWidget {
  final Agent agent;

  const AgentEditDialog({required this.agent, super.key});

  @override
  State<AgentEditDialog> createState() => _AgentEditDialogState();
}

class _AgentEditDialogState extends State<AgentEditDialog> {
  late final TextEditingController _host;
  late final TextEditingController _port;
  late final TextEditingController _metricsPort;
  late final TextEditingController _status;
  late bool _enabled;

  @override
  void initState() {
    super.initState();
    _host = TextEditingController(text: widget.agent.host);
    _port = TextEditingController(text: '${widget.agent.port}');
    _metricsPort = TextEditingController(text: '${widget.agent.metricsPort}');
    _status = TextEditingController(text: widget.agent.status);
    _enabled = widget.agent.enabled;
  }

  @override
  void dispose() {
    _host.dispose();
    _port.dispose();
    _metricsPort.dispose();
    _status.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Edit agent'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: _host,
              decoration: const InputDecoration(labelText: 'Host'),
            ),
            const SizedBox(height: 14),
            TextField(
              controller: _port,
              decoration: const InputDecoration(labelText: 'Port'),
              keyboardType: TextInputType.number,
            ),
            const SizedBox(height: 14),
            TextField(
              controller: _metricsPort,
              decoration: const InputDecoration(labelText: 'Metrics port'),
              keyboardType: TextInputType.number,
            ),
            const SizedBox(height: 14),
            // Status is reported by the agent itself, so it is shown
            // read-only. Only host, ports and the enabled flag are
            // operator-controlled.
            TextField(
              controller: _status,
              enabled: false,
              decoration: const InputDecoration(
                labelText: 'Status',
                helperText: 'Reported by the agent',
              ),
            ),
            const SizedBox(height: 6),
            SwitchListTile(
              title: const Text('Enabled'),
              value: _enabled,
              onChanged: (value) => setState(() => _enabled = value),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: () {
            // `status` is deliberately not assigned: it is reported by the
            // agent, and deepCopy() already carries the current value.
            final updated = widget.agent.deepCopy()
              ..host = _host.text
              ..port = int.tryParse(_port.text) ?? widget.agent.port
              ..metricsPort =
                  int.tryParse(_metricsPort.text) ?? widget.agent.metricsPort
              ..enabled = _enabled;

            Navigator.of(context).pop(updated);
          },
          child: const Text('Save'),
        ),
      ],
    );
  }
}
