import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import 'auth.dart';
import 'data/repo.dart';
import 'data/service.dart';
import 'data/token.dart';
import 'screens/agents_screen.dart';
import 'screens/home_shell.dart';
import 'screens/login.dart';
import 'screens/register_user.dart';
import 'screens/targets_screen.dart';
import 'screens/tests_screen.dart';
import 'screens/use_cases_screen.dart';
import 'screens/users_screen.dart';

class ChaosMasterApp extends StatefulWidget {
  const ChaosMasterApp({super.key});

  @override
  State<ChaosMasterApp> createState() => _ChaosMasterAppState();
}

class _ChaosMasterAppState extends State<ChaosMasterApp> {
  late final TokenStore _tokenStore;
  late final Backend _backend;
  late final AuthService _authService;
  late final AgentsRepository _agentsRepository;
  late final TargetsRepository _targetsRepository;
  late final UseCasesRepository _useCasesRepository;
  late final ExecutionsRepository _executionsRepository;
  late final UsersRepository _usersRepository;

  @override
  void initState() {
    super.initState();

    _tokenStore = TokenStore();
    _backend = Backend(tokenStore: _tokenStore);
    _authService = AuthService(backend: _backend, tokenStore: _tokenStore);

    _agentsRepository = AgentsRepository(backend: _backend);
    _targetsRepository = TargetsRepository(backend: _backend);
    _useCasesRepository = UseCasesRepository(backend: _backend);
    _executionsRepository = ExecutionsRepository(backend: _backend);
    _usersRepository = UsersRepository(backend: _backend);
  }

  @override
  void dispose() {
    _agentsRepository.dispose();
    _targetsRepository.dispose();
    _useCasesRepository.dispose();
    _executionsRepository.dispose();
    _usersRepository.dispose();
    _backend.shutdown();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Chaos Master',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(colorSchemeSeed: Colors.deepPurple, useMaterial3: true),
      builder: (context, child) {
        if (child == null) {
          throw StateError('No child widget provided to MaterialApp.router');
        }
        return AuthScope(notifier: _authService, child: child);
      },
      routerConfig: GoRouter(
        refreshListenable: _authService,
        initialLocation: '/agents',
        redirect: (context, state) {
          final onAuthPage =
              state.fullPath == '/login' || state.fullPath == '/register';

          if (!_authService.isAuthenticated && !onAuthPage) {
            return '/login';
          }
          if (_authService.isAuthenticated && onAuthPage) {
            return '/agents';
          }
          return null;
        },
        routes: [
          GoRoute(
            path: '/login',
            builder: (context, state) => const LoginScreen(),
          ),
          GoRoute(
            path: '/register',
            builder: (context, state) => const RegisterUserScreen(),
          ),
          ShellRoute(
            builder: (context, state, child) => HomeShell(child: child),
            routes: [
              GoRoute(
                path: '/agents',
                builder: (context, state) =>
                    AgentsScreen(repository: _agentsRepository),
              ),
              GoRoute(
                path: '/targets',
                builder: (context, state) =>
                    TargetsScreen(repository: _targetsRepository),
              ),
              GoRoute(
                path: '/use-cases',
                builder: (context, state) =>
                    UseCasesScreen(repository: _useCasesRepository),
              ),
              GoRoute(
                path: '/tests',
                builder: (context, state) => TestsScreen(
                  executions: _executionsRepository,
                  targets: _targetsRepository,
                  useCases: _useCasesRepository,
                  agents: _agentsRepository,
                ),
              ),
              GoRoute(
                path: '/users',
                builder: (context, state) =>
                    UsersScreen(repository: _usersRepository),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
