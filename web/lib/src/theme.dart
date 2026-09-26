import 'package:flutter/material.dart';

/// Palette for the **Aether Flight Deck** ("Deep Space Cyan") design language.
///
/// The whole scheme is derived from a single accent family (Quantum Cyan) on a
/// near-black aerospace canvas. Semantic accents are used only for status:
/// success/nominal, warning/jitter and critical/abort.
class AetherPalette {
  AetherPalette._();

  // Canvas & surfaces
  static const canvas = Color(0xFF0A0E18);
  static const surface = Color(0xFF0F131D);
  static const surfaceHigh = Color(0xFF141A26);
  static const surfaceHighest = Color(0xFF1A2130);

  // Primary accent: Quantum Cyan
  static const cyan = Color(0xFF06B6D4);
  static const cyanBright = Color(0xFF22D3EE);
  static const cyanGlow = Color(0xFF00F3FF);

  // Success / nominal: Emerald Phosphor
  static const emerald = Color(0xFF10B981);
  static const emeraldBright = Color(0xFF00FF88);

  // Warning / jitter: Amber Solar Flare
  static const amber = Color(0xFFF59E0B);

  // Critical / abort: Crimson Nova
  static const crimson = Color(0xFFEF4444);
  static const crimsonDeep = Color(0xFFDC2626);

  // Translucent cyan micro-borders
  static const border = Color(0x3306B6D4);
  static const borderSoft = Color(0x1F06B6D4);

  // Text
  static const textPrimary = Color(0xFFE6F7FB);
  static const textMuted = Color(0xFF8FA3B0);

  /// Ink used on top of the bright cyan accent.
  static const onAccent = Color(0xFF04141A);
}

/// Typographic helpers.
///
/// Headings use the platform UI font; anything numeric/identifier-like
/// (ports, ids, rates, timestamps) uses [monoFamily] so columns align.
/// On Flutter web `monospace` resolves through CSS to the platform's
/// monospace family, so no font asset is required.
class AetherText {
  AetherText._();

  static const monoFamily = 'monospace';

  static TextStyle mono({
    double size = 13,
    FontWeight weight = FontWeight.w500,
    Color color = AetherPalette.textPrimary,
    double letterSpacing = 0.2,
  }) => TextStyle(
    fontFamily: monoFamily,
    fontSize: size,
    fontWeight: weight,
    color: color,
    letterSpacing: letterSpacing,
  );

  /// Small uppercase label used for panel headers, KPI labels and eyebrows.
  static const sectionLabel = TextStyle(
    fontSize: 11,
    fontWeight: FontWeight.w700,
    letterSpacing: 1.8,
    color: AetherPalette.textMuted,
  );

  /// Screen-title treatment: uppercase with wide tracking.
  static const screenTitle = TextStyle(
    fontSize: 15,
    fontWeight: FontWeight.w700,
    letterSpacing: 1.8,
    color: AetherPalette.textPrimary,
  );
}

/// Shared decorations (borders, glows, panels).
class AetherDecor {
  AetherDecor._();

  static BorderRadius get radius => BorderRadius.circular(14);

  static BoxDecoration panel({Color? accent}) => BoxDecoration(
    color: AetherPalette.surface,
    borderRadius: radius,
    border: Border.all(color: AetherPalette.border),
    boxShadow: [
      BoxShadow(
        color: (accent ?? AetherPalette.cyan).withValues(alpha: 0.07),
        blurRadius: 24,
        spreadRadius: -8,
      ),
    ],
  );

  /// Soft accent glow used behind KPI values and status dots.
  static List<BoxShadow> glow(
    Color color, {
    double alpha = 0.28,
    double blur = 16,
  }) => [
    BoxShadow(
      color: color.withValues(alpha: alpha),
      blurRadius: blur,
      spreadRadius: -4,
    ),
  ];

  /// Quantum Cyan mixed into the canvas, used as the opaque centre of the
  /// backdrop bloom.
  static const canvasBloom = Color(0xFF0E2A36);

  /// Full-viewport aerospace backdrop: flat canvas with a single soft
  /// radial cyan bloom in the upper-left.
  ///
  /// Both stops are **opaque** on purpose. The browser canvas behind the app
  /// is transparent, so a semi-transparent stop would let the page background
  /// bleed through and wash the console out.
  static const backdropGradient = RadialGradient(
    center: Alignment(-0.85, -0.95),
    radius: 1.7,
    colors: [canvasBloom, AetherPalette.canvas],
  );
}

/// Wraps a subtree in the aerospace backdrop.
///
/// The base colour is painted by a separate [ColoredBox] rather than as the
/// `color` of a `BoxDecoration` that also has a gradient, so the dark canvas
/// is guaranteed to be laid down even if the gradient has any transparency.
class AetherBackground extends StatelessWidget {
  const AetherBackground({required this.child, super.key});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: AetherPalette.canvas,
      child: DecoratedBox(
        decoration: const BoxDecoration(gradient: AetherDecor.backdropGradient),
        child: child,
      ),
    );
  }
}

OutlineInputBorder _inputBorder(Color color, {double width = 1}) =>
    OutlineInputBorder(
      borderRadius: BorderRadius.circular(10),
      borderSide: BorderSide(color: color, width: width),
    );

/// Builds the single dark theme used across the app.
ThemeData aetherTheme() {
  const scheme = ColorScheme.dark(
    primary: AetherPalette.cyan,
    onPrimary: AetherPalette.onAccent,
    primaryContainer: Color(0xFF0E3A46),
    onPrimaryContainer: AetherPalette.cyanBright,
    secondary: AetherPalette.cyanBright,
    onSecondary: AetherPalette.onAccent,
    secondaryContainer: Color(0xFF10303A),
    onSecondaryContainer: AetherPalette.cyanGlow,
    tertiary: AetherPalette.emerald,
    onTertiary: AetherPalette.onAccent,
    error: AetherPalette.crimson,
    onError: Colors.white,
    errorContainer: Color(0xFF3A1416),
    onErrorContainer: Color(0xFFFFD9DA),
    surface: AetherPalette.surface,
    onSurface: AetherPalette.textPrimary,
    onSurfaceVariant: AetherPalette.textMuted,
    surfaceContainerHighest: AetherPalette.surfaceHighest,
    outline: AetherPalette.border,
    outlineVariant: AetherPalette.borderSoft,
    shadow: Colors.black,
    scrim: Colors.black,
    inverseSurface: AetherPalette.textPrimary,
    onInverseSurface: AetherPalette.canvas,
    inversePrimary: Color(0xFF0E7490),
  );

  final base = ThemeData(
    useMaterial3: true,
    brightness: Brightness.dark,
    colorScheme: scheme,
    visualDensity: VisualDensity.compact,
  );

  final textTheme = base.textTheme.apply(
    bodyColor: AetherPalette.textPrimary,
    displayColor: AetherPalette.textPrimary,
  );

  return base.copyWith(
    scaffoldBackgroundColor: AetherPalette.canvas,
    canvasColor: AetherPalette.surface,
    textTheme: textTheme,
    dividerTheme: const DividerThemeData(
      color: AetherPalette.border,
      thickness: 1,
      space: 1,
    ),
    textSelectionTheme: TextSelectionThemeData(
      cursorColor: AetherPalette.cyanBright,
      selectionColor: AetherPalette.cyan.withValues(alpha: 0.30),
      selectionHandleColor: AetherPalette.cyan,
    ),
    appBarTheme: AppBarTheme(
      backgroundColor: AetherPalette.surface.withValues(alpha: 0.72),
      surfaceTintColor: Colors.transparent,
      foregroundColor: AetherPalette.textPrimary,
      elevation: 0,
      scrolledUnderElevation: 0,
      toolbarHeight: 64,
      centerTitle: false,
      titleTextStyle: AetherText.screenTitle,
      iconTheme: const IconThemeData(color: AetherPalette.textMuted, size: 20),
      actionsIconTheme: const IconThemeData(
        color: AetherPalette.textMuted,
        size: 20,
      ),
      shape: const Border(bottom: BorderSide(color: AetherPalette.border)),
    ),
    navigationRailTheme: NavigationRailThemeData(
      backgroundColor: AetherPalette.surface.withValues(alpha: 0.55),
      elevation: 0,
      useIndicator: true,
      indicatorColor: AetherPalette.cyan.withValues(alpha: 0.16),
      indicatorShape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(10),
        side: const BorderSide(color: AetherPalette.border),
      ),
      selectedIconTheme: const IconThemeData(
        color: AetherPalette.cyanBright,
        size: 22,
      ),
      unselectedIconTheme: const IconThemeData(
        color: AetherPalette.textMuted,
        size: 22,
      ),
      selectedLabelTextStyle: const TextStyle(
        color: AetherPalette.cyanBright,
        fontSize: 11,
        fontWeight: FontWeight.w700,
        letterSpacing: 0.8,
      ),
      unselectedLabelTextStyle: const TextStyle(
        color: AetherPalette.textMuted,
        fontSize: 11,
        fontWeight: FontWeight.w600,
        letterSpacing: 0.8,
      ),
      labelType: NavigationRailLabelType.all,
      minWidth: 92,
    ),
    dataTableTheme: DataTableThemeData(
      headingRowColor: WidgetStatePropertyAll(
        AetherPalette.surfaceHigh.withValues(alpha: 0.85),
      ),
      headingTextStyle: const TextStyle(
        color: AetherPalette.cyanBright,
        fontSize: 11,
        fontWeight: FontWeight.w700,
        letterSpacing: 1.4,
      ),
      dataTextStyle: const TextStyle(
        color: AetherPalette.textPrimary,
        fontSize: 13,
      ),
      headingRowHeight: 46,
      dataRowMinHeight: 46,
      dataRowMaxHeight: 58,
      horizontalMargin: 18,
      columnSpacing: 30,
      dividerThickness: 1,
    ),
    inputDecorationTheme: InputDecorationThemeData(
      isDense: true,
      filled: true,
      fillColor: AetherPalette.canvas.withValues(alpha: 0.55),
      contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
      labelStyle: const TextStyle(
        color: AetherPalette.textMuted,
        letterSpacing: 0.4,
      ),
      floatingLabelStyle: const TextStyle(
        color: AetherPalette.cyanBright,
        letterSpacing: 0.4,
      ),
      hintStyle: TextStyle(
        color: AetherPalette.textMuted.withValues(alpha: 0.55),
        letterSpacing: 0.3,
      ),
      errorStyle: const TextStyle(
        color: AetherPalette.crimson,
        letterSpacing: 0.3,
      ),
      border: _inputBorder(AetherPalette.border),
      enabledBorder: _inputBorder(AetherPalette.border),
      focusedBorder: _inputBorder(AetherPalette.cyan, width: 1.4),
      errorBorder: _inputBorder(AetherPalette.crimson),
      focusedErrorBorder: _inputBorder(AetherPalette.crimson, width: 1.4),
      prefixIconColor: AetherPalette.textMuted,
      suffixIconColor: AetherPalette.textMuted,
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: AetherPalette.cyan,
        foregroundColor: AetherPalette.onAccent,
        disabledBackgroundColor: AetherPalette.cyan.withValues(alpha: 0.22),
        disabledForegroundColor: AetherPalette.textMuted,
        elevation: 0,
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        textStyle: const TextStyle(
          fontWeight: FontWeight.w700,
          letterSpacing: 1.0,
          fontSize: 13,
        ),
      ),
    ),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: AetherPalette.cyanBright,
        side: const BorderSide(color: AetherPalette.border),
        padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 14),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        textStyle: const TextStyle(
          fontWeight: FontWeight.w700,
          letterSpacing: 0.8,
          fontSize: 12.5,
        ),
      ),
    ),
    textButtonTheme: TextButtonThemeData(
      style: TextButton.styleFrom(
        foregroundColor: AetherPalette.cyanBright,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        textStyle: const TextStyle(
          fontWeight: FontWeight.w600,
          letterSpacing: 0.6,
          fontSize: 12.5,
        ),
      ),
    ),
    iconButtonTheme: IconButtonThemeData(
      style: ButtonStyle(
        foregroundColor: WidgetStateProperty.resolveWith<Color?>(
          (states) => states.contains(WidgetState.hovered)
              ? AetherPalette.cyanBright
              : AetherPalette.textMuted,
        ),
        overlayColor: WidgetStatePropertyAll(
          AetherPalette.cyan.withValues(alpha: 0.10),
        ),
      ),
    ),
    floatingActionButtonTheme: FloatingActionButtonThemeData(
      backgroundColor: AetherPalette.cyan,
      foregroundColor: AetherPalette.onAccent,
      elevation: 0,
      focusElevation: 0,
      hoverElevation: 0,
      highlightElevation: 0,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
    ),
    snackBarTheme: SnackBarThemeData(
      backgroundColor: AetherPalette.surfaceHighest,
      contentTextStyle: const TextStyle(
        color: AetherPalette.textPrimary,
        fontSize: 13,
      ),
      actionTextColor: AetherPalette.cyanBright,
      behavior: SnackBarBehavior.floating,
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(10),
        side: const BorderSide(color: AetherPalette.border),
      ),
    ),
    chipTheme: ChipThemeData(
      backgroundColor: AetherPalette.cyan.withValues(alpha: 0.12),
      side: const BorderSide(color: AetherPalette.border),
      labelStyle: AetherText.mono(
        size: 11,
        weight: FontWeight.w700,
        color: AetherPalette.cyanBright,
        letterSpacing: 1.2,
      ),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      showCheckmark: false,
    ),
    dialogTheme: DialogThemeData(
      backgroundColor: AetherPalette.surface,
      surfaceTintColor: Colors.transparent,
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: const BorderSide(color: AetherPalette.border),
      ),
      titleTextStyle: const TextStyle(
        color: AetherPalette.textPrimary,
        fontSize: 15,
        fontWeight: FontWeight.w700,
        letterSpacing: 1.4,
      ),
      contentTextStyle: const TextStyle(
        color: AetherPalette.textPrimary,
        fontSize: 13.5,
        height: 1.45,
      ),
    ),
    cardTheme: CardThemeData(
      color: AetherPalette.surface,
      surfaceTintColor: Colors.transparent,
      elevation: 0,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: AetherDecor.radius,
        side: const BorderSide(color: AetherPalette.border),
      ),
    ),
    listTileTheme: const ListTileThemeData(
      iconColor: AetherPalette.textMuted,
      textColor: AetherPalette.textPrimary,
      titleTextStyle: TextStyle(
        color: AetherPalette.textPrimary,
        fontSize: 14,
        fontWeight: FontWeight.w600,
      ),
      subtitleTextStyle: TextStyle(
        color: AetherPalette.textMuted,
        fontSize: 12.5,
        height: 1.5,
      ),
      contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
    ),
    switchTheme: SwitchThemeData(
      thumbColor: WidgetStateProperty.resolveWith<Color?>(
        (states) => states.contains(WidgetState.selected)
            ? AetherPalette.onAccent
            : AetherPalette.textMuted,
      ),
      trackColor: WidgetStateProperty.resolveWith<Color?>(
        (states) => states.contains(WidgetState.selected)
            ? AetherPalette.cyan
            : AetherPalette.surfaceHighest,
      ),
      trackOutlineColor: const WidgetStatePropertyAll(AetherPalette.border),
    ),
    progressIndicatorTheme: const ProgressIndicatorThemeData(
      color: AetherPalette.cyan,
      linearTrackColor: AetherPalette.surfaceHighest,
      circularTrackColor: AetherPalette.surfaceHighest,
    ),
    tooltipTheme: TooltipThemeData(
      waitDuration: const Duration(milliseconds: 300),
      textStyle: const TextStyle(
        color: AetherPalette.textPrimary,
        fontSize: 11.5,
        letterSpacing: 0.4,
      ),
      decoration: BoxDecoration(
        color: AetherPalette.surfaceHighest,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AetherPalette.border),
      ),
    ),
  );
}
