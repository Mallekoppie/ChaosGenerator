import 'package:flutter/material.dart';

import '../theme.dart';

/// Console-styled data table.
///
/// Scrolls horizontally when the viewport is narrower than the table so that
/// no column is ever clipped, and stretches to the available width otherwise.
class AetherDataTable extends StatelessWidget {
  const AetherDataTable({required this.columns, required this.rows, super.key});

  final List<DataColumn> columns;
  final List<DataRow> rows;

  /// Standard row hover highlight for console tables.
  static WidgetStateProperty<Color?>? get rowHighlight =>
      WidgetStateProperty.resolveWith<Color?>(
        (states) => states.contains(WidgetState.hovered)
            ? AetherPalette.cyan.withValues(alpha: 0.05)
            : null,
      );

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: ConstrainedBox(
            constraints: BoxConstraints(
              minWidth: constraints.maxWidth.isFinite
                  ? constraints.maxWidth
                  : 0,
            ),
            child: DataTable(
              columns: columns,
              rows: rows,
              showCheckboxColumn: false,
              border: const TableBorder(
                horizontalInside: BorderSide(color: AetherPalette.borderSoft),
              ),
            ),
          ),
        );
      },
    );
  }
}
