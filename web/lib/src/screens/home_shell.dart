import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../auth.dart';
import '../theme.dart';
import '../widgets/aether_chrome.dart';

class HomeShell extends StatefulWidget {
  final Widget child;

  /// Current route path, used to keep the nav rail selection in sync.
  final String location;

  const HomeShell({required this.child, required this.location, super.key});

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  static const _destinations = <_NavDestination>[
    _NavDestination(path: '/sys-cm', icon: Icons.bolt, label: 'SYS-CM'),
    _NavDestination(path: '/agents', icon: Icons.dns, label: 'Agents'),
    _NavDestination(
      path: '/targets',
      icon: Icons.my_location,
      label: 'Targets',
    ),
    _NavDestination(
      path: '/use-cases',
      icon: Icons.list_alt,
      label: 'Use Cases',
    ),
    _NavDestination(path: '/tests', icon: Icons.play_circle, label: 'Tests'),
    _NavDestination(path: '/users', icon: Icons.people, label: 'Users'),
  ];

  /// Index of the destination matching the active route. Agents is the
  /// post-login default landing screen.
  int get _selectedIndex {
    for (var i = 0; i < _destinations.length; i++) {
      if (widget.location.startsWith(_destinations[i].path)) {
        return i;
      }
    }
    return 1;
  }

  void _select(int index) => context.go(_destinations[index].path);

  @override
  Widget build(BuildContext context) {
    return AetherBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        body: SafeArea(
          child: Column(
            children: [
              const GlobalTopBar(),
              Expanded(
                child: Row(
                  children: [
                    NavigationRail(
                      selectedIndex: _selectedIndex,
                      labelType: NavigationRailLabelType.all,
                      onDestinationSelected: _select,
                      trailing: Padding(
                        padding: const EdgeInsets.only(top: 12),
                        child: IconButton(
                          tooltip: 'Log out',
                          icon: const Icon(Icons.logout),
                          onPressed: () {
                            AuthService.of(context).logout();
                            context.go('/login');
                          },
                        ),
                      ),
                      destinations: [
                        for (final destination in _destinations)
                          NavigationRailDestination(
                            icon: Icon(destination.icon),
                            label: Text(destination.label),
                          ),
                      ],
                    ),
                    const VerticalDivider(thickness: 1, width: 1),
                    Expanded(child: widget.child),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// A single entry in the persistent navigation rail.
class _NavDestination {
  const _NavDestination({
    required this.path,
    required this.icon,
    required this.label,
  });

  final String path;
  final IconData icon;
  final String label;
}
