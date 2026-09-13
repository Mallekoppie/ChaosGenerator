import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../auth.dart';

class HomeShell extends StatefulWidget {
  final Widget child;

  const HomeShell({required this.child, super.key});

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  int _selectedIndex = 0;

  void _select(int index) {
    setState(() {
      _selectedIndex = index;
    });

    switch (index) {
      case 0:
        context.go('/agents');
        break;
      case 1:
        context.go('/targets');
        break;
      case 2:
        context.go('/use-cases');
        break;
      case 3:
        context.go('/tests');
        break;
      case 4:
        context.go('/users');
        break;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Row(
          children: [
            NavigationRail(
              selectedIndex: _selectedIndex,
              labelType: NavigationRailLabelType.all,
              onDestinationSelected: _select,
              trailing: Padding(
                padding: const EdgeInsets.only(top: 8),
                child: IconButton(
                  tooltip: 'Log out',
                  icon: const Icon(Icons.logout),
                  onPressed: () {
                    AuthService.of(context).logout();
                    context.go('/login');
                  },
                ),
              ),
              destinations: const [
                NavigationRailDestination(
                  icon: Icon(Icons.dns),
                  label: Text('Agents'),
                ),
                NavigationRailDestination(
                  icon: Icon(Icons.my_location),
                  label: Text('Targets'),
                ),
                NavigationRailDestination(
                  icon: Icon(Icons.list_alt),
                  label: Text('Use Cases'),
                ),
                NavigationRailDestination(
                  icon: Icon(Icons.play_circle),
                  label: Text('Tests'),
                ),
                NavigationRailDestination(
                  icon: Icon(Icons.people),
                  label: Text('Users'),
                ),
              ],
            ),
            const VerticalDivider(thickness: 1, width: 1),
            Expanded(child: widget.child),
          ],
        ),
      ),
    );
  }
}
