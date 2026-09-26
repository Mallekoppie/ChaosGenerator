import 'package:flutter/material.dart';

import '../theme.dart';

/// Monospace text used for ids, ports, rates and timestamps so that
/// columns of telemetry line up.
class TelemetryText extends StatelessWidget {
  const TelemetryText(
    this.data, {
    this.size = 13,
    this.weight = FontWeight.w500,
    this.color = AetherPalette.textPrimary,
    this.align,
    this.overflow,
    super.key,
  });

  final String data;
  final double size;
  final FontWeight weight;
  final Color color;
  final TextAlign? align;
  final TextOverflow? overflow;

  @override
  Widget build(BuildContext context) {
    return Text(
      data,
      textAlign: align,
      overflow: overflow,
      style: AetherText.mono(size: size, weight: weight, color: color),
    );
  }
}

/// Tiny uppercase eyebrow with a glowing accent bar.
class SectionHeading extends StatelessWidget {
  const SectionHeading({
    required this.label,
    this.trailing,
    this.accent = AetherPalette.cyan,
    super.key,
  });

  final String label;
  final Widget? trailing;
  final Color accent;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          width: 3,
          height: 14,
          decoration: BoxDecoration(
            color: accent,
            borderRadius: BorderRadius.circular(2),
            boxShadow: AetherDecor.glow(accent, alpha: 0.55, blur: 10),
          ),
        ),
        const SizedBox(width: 10),
        Expanded(
          child: Text(
            label.toUpperCase(),
            style: AetherText.sectionLabel,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        if (trailing != null) trailing!,
      ],
    );
  }
}

/// Glowing status dot.
class StatusDot extends StatelessWidget {
  const StatusDot({
    this.color = AetherPalette.emeraldBright,
    this.size = 8,
    super.key,
  });

  final Color color;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: color,
        shape: BoxShape.circle,
        boxShadow: AetherDecor.glow(color, alpha: 0.8, blur: 8),
      ),
    );
  }
}

/// Centred console message used for the empty states.
class ConsoleMessage extends StatelessWidget {
  const ConsoleMessage(this.message, {this.icon, this.color, super.key});

  final String message;
  final IconData? icon;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final tint = color ?? AetherPalette.textMuted;

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (icon != null) ...[
              Icon(icon, size: 26, color: tint),
              const SizedBox(height: 12),
            ],
            Text(
              message,
              textAlign: TextAlign.center,
              style: TextStyle(color: tint, fontSize: 13, letterSpacing: 0.6),
            ),
          ],
        ),
      ),
    );
  }
}

/// Full-screen loading indicator for the first load of a screen.
class ConsoleLoading extends StatelessWidget {
  const ConsoleLoading({super.key});

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Padding(
        padding: EdgeInsets.all(48),
        child: SizedBox(
          height: 34,
          width: 34,
          child: CircularProgressIndicator(strokeWidth: 2.5),
        ),
      ),
    );
  }
}

/// Boolean cell: emerald check when true, muted cross when false.
Widget boolCell(bool value) => Icon(
  value ? Icons.check_circle : Icons.cancel_outlined,
  size: 18,
  color: value ? AetherPalette.emeraldBright : AetherPalette.textMuted,
);

/// Horizontal meter used by the telemetry breakdown lists.
class MeterBar extends StatelessWidget {
  const MeterBar({
    required this.fraction,
    this.color = AetherPalette.cyan,
    super.key,
  });

  final double fraction;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(4),
      child: LinearProgressIndicator(
        value: fraction.clamp(0.0, 1.0),
        minHeight: 6,
        backgroundColor: AetherPalette.surfaceHighest,
        valueColor: AlwaysStoppedAnimation<Color>(color),
      ),
    );
  }
}

/// Maps a free-text agent status onto a semantic accent.
Color statusColor(String status) {
  switch (status.trim().toLowerCase()) {
    case 'online':
    case 'healthy':
    case 'ready':
      return AetherPalette.emeraldBright;
    case 'degraded':
    case 'busy':
      return AetherPalette.amber;
    case 'offline':
    case 'standby':
    case 'unknown':
      return AetherPalette.textMuted;
    default:
      return AetherPalette.cyanBright;
  }
}

/// Maps a target protocol onto a semantic accent.
Color protocolColor(String protocol) {
  switch (protocol.trim().toLowerCase()) {
    case 'https':
      return AetherPalette.emeraldBright;
    case 'grpc':
      return AetherPalette.amber;
    case 'http':
    default:
      return AetherPalette.cyanBright;
  }
}

/// Colour-coded HTTP method badge used by the use-case catalogue.
class MethodChip extends StatelessWidget {
  const MethodChip(this.method, {super.key});

  final String method;

  static Color colorFor(String method) {
    switch (method.toUpperCase()) {
      case 'POST':
        return AetherPalette.emeraldBright;
      case 'PUT':
      case 'PATCH':
        return AetherPalette.amber;
      case 'DELETE':
        return AetherPalette.crimson;
      case 'GET':
      default:
        return AetherPalette.cyanBright;
    }
  }

  @override
  Widget build(BuildContext context) {
    final color = colorFor(method);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withValues(alpha: 0.45)),
        boxShadow: AetherDecor.glow(color, alpha: 0.18, blur: 10),
      ),
      child: Text(
        method.toUpperCase(),
        style: AetherText.mono(
          size: 11,
          weight: FontWeight.w700,
          color: color,
          letterSpacing: 1.2,
        ),
      ),
    );
  }
}
