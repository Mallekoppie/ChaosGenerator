import 'package:flutter/material.dart';

import '../theme.dart';
import 'aether_common.dart';

/// Brand tagline shown in the global top bar.
const aetherTagline =
    'Professionally breaking things before your users do — Orchestrating '
    'high-entropy failure so production never has to panic.';

/// Compact rounded status pill (`GRID: NOMINAL`, `LATENCY SLA: 99.95%`, ...).
class StatusPill extends StatelessWidget {
  const StatusPill({
    required this.label,
    this.color = AetherPalette.cyanBright,
    this.icon,
    super.key,
  });

  final String label;
  final Color color;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.10),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: color.withValues(alpha: 0.35)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (icon != null) ...[
            Icon(icon, size: 12, color: color),
            const SizedBox(width: 6),
          ] else ...[
            StatusDot(color: color, size: 6),
            const SizedBox(width: 6),
          ],
          Text(
            label.toUpperCase(),
            style: AetherText.mono(
              size: 10,
              weight: FontWeight.w700,
              color: color,
              letterSpacing: 1.0,
            ),
          ),
        ],
      ),
    );
  }
}

/// Permanent, full-viewport top bar: brand, tagline and global status pills.
class GlobalTopBar extends StatelessWidget {
  const GlobalTopBar({super.key});

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final showTagline = constraints.maxWidth > 1000;
        final showPills = constraints.maxWidth > 720;

        return Container(
          height: 54,
          padding: const EdgeInsets.symmetric(horizontal: 18),
          decoration: const BoxDecoration(
            border: Border(bottom: BorderSide(color: AetherPalette.border)),
            gradient: LinearGradient(
              begin: Alignment.centerLeft,
              end: Alignment.centerRight,
              colors: [AetherDecor.canvasBloom, AetherPalette.surface],
            ),
          ),
          child: Row(
            children: [
              Icon(
                Icons.bolt,
                size: 20,
                color: AetherPalette.cyanGlow,
                shadows: AetherDecor.glow(AetherPalette.cyanGlow, alpha: 0.9),
              ),
              const SizedBox(width: 10),
              const Text(
                'CHAOSPROCESSOR',
                style: TextStyle(
                  color: AetherPalette.textPrimary,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 3.0,
                ),
              ),
              if (showTagline) ...[
                const SizedBox(width: 20),
                const Flexible(
                  child: Text(
                    aetherTagline,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: AetherPalette.textMuted,
                      fontSize: 11.5,
                      letterSpacing: 0.3,
                    ),
                  ),
                ),
              ],
              const SizedBox(width: 20),
              if (showPills) ...[
                const StatusPill(
                  label: 'GRID: NOMINAL',
                  color: AetherPalette.emeraldBright,
                  icon: Icons.check_circle,
                ),
                const SizedBox(width: 8),
              ],
              const StatusPill(label: 'CONTROLLED ENTROPY'),
            ],
          ),
        );
      },
    );
  }
}

/// Live operational pills shown in the secondary (per-screen) bar.
class OpsStatusPills extends StatelessWidget {
  const OpsStatusPills({super.key});

  @override
  Widget build(BuildContext context) {
    return const Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        StatusPill(
          label: 'LATENCY SLA: 99.95%',
          color: AetherPalette.emeraldBright,
        ),
        SizedBox(width: 8),
        StatusPill(label: 'ORBIT_POS: GEO-STATIONARY'),
      ],
    );
  }
}

/// Signed-in operator badge.
class OperatorProfile extends StatelessWidget {
  const OperatorProfile({super.key});

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: 'Signed-in operator',
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 8),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              height: 28,
              width: 28,
              decoration: BoxDecoration(
                color: AetherPalette.cyan.withValues(alpha: 0.16),
                shape: BoxShape.circle,
                border: Border.all(color: AetherPalette.border),
              ),
              child: const Icon(
                Icons.person,
                size: 16,
                color: AetherPalette.cyanBright,
              ),
            ),
            const SizedBox(width: 8),
            const Text(
              'OPERATOR',
              style: TextStyle(
                color: AetherPalette.textMuted,
                fontSize: 10.5,
                fontWeight: FontWeight.w700,
                letterSpacing: 1.4,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Secondary operational sub-bar for a screen.
///
/// Shows the screen name, the live SLA/node pills (on wide viewports), the
/// screen's own actions and the operator badge. Replaces the plain app bar.
PreferredSizeWidget aetherAppBar(
  BuildContext context, {
  required String title,
  List<Widget> actions = const [],
}) {
  final wide = MediaQuery.sizeOf(context).width >= 1280;

  return AppBar(
    title: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        const StatusDot(color: AetherPalette.cyanBright, size: 7),
        const SizedBox(width: 10),
        Text(title.toUpperCase(), style: AetherText.screenTitle),
      ],
    ),
    actions: [
      if (wide) ...[const OpsStatusPills(), const SizedBox(width: 14)],
      ...actions,
      const SizedBox(width: 4),
      const OperatorProfile(),
      const SizedBox(width: 10),
    ],
  );
}
