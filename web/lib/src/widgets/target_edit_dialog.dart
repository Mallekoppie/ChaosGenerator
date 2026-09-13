import 'package:flutter/material.dart';

import '../generated/generated.dart';

/// Dialog used to add or edit a test target.
class TargetEditDialog extends StatefulWidget {
  /// The target being edited, or null when adding a new one.
  final Target? target;

  const TargetEditDialog({this.target, super.key});

  @override
  State<TargetEditDialog> createState() => _TargetEditDialogState();
}

class _TargetEditDialogState extends State<TargetEditDialog> {
  static const _protocols = ['http', 'https', 'grpc'];

  late final TextEditingController _name;
  late final TextEditingController _address;
  late final TextEditingController _sni;
  late String _protocol;
  late bool _connectionReuse;

  @override
  void initState() {
    super.initState();
    final target = widget.target;

    _name = TextEditingController(text: target?.name ?? '');
    _address = TextEditingController(text: target?.address ?? '');
    _sni = TextEditingController(text: target?.sni ?? '');
    _protocol = target?.protocol.isNotEmpty == true ? target!.protocol : 'http';
    _connectionReuse = target?.connectionReuseEnabled ?? false;
  }

  @override
  void dispose() {
    _name.dispose();
    _address.dispose();
    _sni.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(widget.target == null ? 'Add target' : 'Edit target'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: _name,
              decoration: const InputDecoration(labelText: 'Name'),
            ),
            TextField(
              controller: _address,
              decoration: const InputDecoration(
                labelText: 'Address',
                hintText: 'http://localhost:8080',
              ),
            ),
            DropdownButtonFormField<String>(
              initialValue: _protocol,
              decoration: const InputDecoration(labelText: 'Protocol'),
              items: _protocols
                  .map(
                    (protocol) => DropdownMenuItem(
                      value: protocol,
                      child: Text(protocol),
                    ),
                  )
                  .toList(),
              onChanged: (value) {
                if (value != null) {
                  setState(() => _protocol = value);
                }
              },
            ),
            TextField(
              controller: _sni,
              decoration: const InputDecoration(labelText: 'SNI (optional)'),
            ),
            SwitchListTile(
              title: const Text('Connection reuse'),
              value: _connectionReuse,
              onChanged: (value) => setState(() => _connectionReuse = value),
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
            final target = (widget.target?.deepCopy() ?? Target())
              ..name = _name.text.trim()
              ..address = _address.text.trim()
              ..protocol = _protocol
              ..sni = _sni.text.trim()
              ..connectionReuseEnabled = _connectionReuse;

            Navigator.of(context).pop(target);
          },
          child: const Text('Save'),
        ),
      ],
    );
  }
}
