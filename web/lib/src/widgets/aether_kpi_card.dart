import 'package:flutter/material.dart';

import '../theme.dart';
import 'aether_common.dart';

/// Telemetry KPI tile: uppercase label, big monospace value with unit,
/// and an optional status hint.
class KpiCard extends StatelessWidget {
  const KpiCard({
    required this.label,
    required this.value,
    this.unit,
    this.hint,
    this.icon,
    this.accent = AetherPalette.cyan,
    this.width = 250,
    super.key,
  });

  final String label;
  final String value;
  final String? unit;
  final String? hint;
  final IconData? icon;
  final Color accent;
  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 16),
      decoration: AetherDecor.panel(accent: accent),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  label.toUpperCase(),
                  style: AetherText.sectionLabel,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (icon != null) Icon(icon, size: 16, color: accent),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            crossAxisAlignment: CrossAxisAlignment.baseline,
            textBaseline: TextBaseline.alphabetic,
            children: [
              Flexible(
                child: TelemetryText(
                  value,
                  size: 22,
                  weight: FontWeight.w700,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (unit != null) ...[
                const SizedBox(width: 6),
                TelemetryText(
                  unit!,
                  size: 11,
                  weight: FontWeight.w600,
                  color: accent,
                ),
              ],
            ],
          ),
          if (hint != null) ...[
            const SizedBox(height: 10),
            Row(
              children: [
                StatusDot(color: accent, size: 6),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    hint!,
                    style: const TextStyle(
                      color: AetherPalette.textMuted,
                      fontSize: 11.5,
                      letterSpacing: 0.3,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }
}

/// Evenly-spaced row of [KpiCard]s that wraps on narrower viewports.
class KpiDeck extends StatelessWidget {
  const KpiDeck({required this.cards, super.key});

  final List<Widget> cards;

  @override
  Widget build(BuildContext context) {
    return Wrap(spacing: 16, runSpacing: 16, children: cards);
  }
}
