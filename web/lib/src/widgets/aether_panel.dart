import 'package:flutter/material.dart';

import '../theme.dart';
import 'aether_common.dart';

/// Bordered, glowing console panel with an optional uppercase section header.
class AetherPanel extends StatelessWidget {
  const AetherPanel({
    required this.child,
    this.title,
    this.trailing,
    this.accent = AetherPalette.cyan,
    this.padding = const EdgeInsets.all(16),
    this.fill = false,
    super.key,
  });

  final Widget child;

  /// Uppercase eyebrow shown in the panel header.
  final String? title;

  /// Optional widget pinned to the right of the header.
  final Widget? trailing;

  /// Accent used for the header bar and the panel glow.
  final Color accent;

  final EdgeInsetsGeometry padding;

  /// When true the panel fills its parent's height and scrolls its child,
  /// which is what the live tables need.
  final bool fill;

  @override
  Widget build(BuildContext context) {
    final header = title == null
        ? const <Widget>[]
        : <Widget>[
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 14, 12, 12),
              child: SectionHeading(
                label: title!,
                accent: accent,
                trailing: trailing,
              ),
            ),
            const Divider(height: 1, color: AetherPalette.borderSoft),
          ];

    return Container(
      decoration: AetherDecor.panel(accent: accent),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(13),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          mainAxisSize: fill ? MainAxisSize.max : MainAxisSize.min,
          children: [
            ...header,
            if (fill)
              Expanded(
                child: Padding(padding: padding, child: child),
              )
            else
              Padding(padding: padding, child: child),
          ],
        ),
      ),
    );
  }
}
