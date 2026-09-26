import 'package:flutter/material.dart';

import '../data/repo.dart';
import '../generated/generated.dart';
import '../theme.dart';
import '../widgets/aether_chrome.dart';
import '../widgets/aether_common.dart';

class UseCasesScreen extends StatefulWidget {
  final UseCasesRepository repository;

  const UseCasesScreen({required this.repository, super.key});

  @override
  State<UseCasesScreen> createState() => _UseCasesScreenState();
}

class _UseCasesScreenState extends State<UseCasesScreen> {
  @override
  void initState() {
    super.initState();
    widget.repository.load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: aetherAppBar(
        context,
        title: 'Use Cases',
        actions: [
          IconButton(
            tooltip: 'Refresh',
            icon: const Icon(Icons.refresh),
            onPressed: widget.repository.load,
          ),
        ],
      ),
      body: ListenableBuilder(
        listenable: widget.repository,
        builder: (context, _) {
          final repository = widget.repository;

          if (repository.loading && repository.useCases.isEmpty) {
            return const ConsoleLoading();
          }

          if (repository.error != null && repository.useCases.isEmpty) {
            return ConsoleMessage(
              repository.error!,
              icon: Icons.error_outline,
              color: AetherPalette.crimson,
            );
          }

          if (repository.useCases.isEmpty) {
            return const ConsoleMessage(
              'No use cases are available',
              icon: Icons.list_alt_outlined,
            );
          }

          return ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: repository.useCases.length,
            separatorBuilder: (context, index) => const SizedBox(height: 12),
            itemBuilder: (context, index) =>
                _UseCaseCard(useCase: repository.useCases[index]),
          );
        },
      ),
    );
  }
}

/// Scenario card: colour-coded method badge, title line, endpoint path and
/// description.
class _UseCaseCard extends StatelessWidget {
  const _UseCaseCard({required this.useCase});

  final UseCase useCase;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: AetherDecor.panel(),
      padding: const EdgeInsets.all(16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          MethodChip(useCase.method),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '${useCase.name}  (${useCase.id})',
                  style: const TextStyle(
                    color: AetherPalette.textPrimary,
                    fontSize: 14,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 0.3,
                  ),
                ),
                const SizedBox(height: 8),
                TelemetryText(
                  useCase.path,
                  size: 12.5,
                  color: AetherPalette.cyanBright,
                ),
                if (useCase.description.isNotEmpty) ...[
                  const SizedBox(height: 6),
                  Text(
                    useCase.description,
                    style: const TextStyle(
                      color: AetherPalette.textMuted,
                      fontSize: 12.5,
                      height: 1.5,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}
