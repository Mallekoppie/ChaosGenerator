# Product Requirements Document (PRD) & Engineering Brief: ChaosProcessor

**Document Version:** 2.0.0  
**Status:** Approved / Active Specification  
**Design System Reference:** `Aether Flight Deck` (Deep Space Cyan)  
**Target Platform:** Desktop Web Application (SRE / Performance Engineering Mission Control)  
**Primary Brand Identity:** **ChaosProcessor**  
**Brand Slogan:** *"Professionally breaking things before your users do — Orchestrating high-entropy failure so production never has to panic."*

---

## 1. Executive Summary & Product Vision

### 1.1 Mission & Value Proposition
**ChaosProcessor** (formerly *Chaos Master*) is an enterprise-grade distributed load generation, synthetic probing, and chaos engineering orchestration control deck. It empowers Site Reliability Engineers (SREs), Infrastructure Architects, and Performance Engineers to subject mission-critical internal and external services to extreme high-entropy conditions, synthetic stress, and network fault injection before outages can manifest in production.

Unlike generic marketing-heavy APMs, ChaosProcessor is an **information-dense administrative flight deck** optimized for immediate situational awareness, rapid mitigation, high-speed telemetry scanning, and zero-latency operational control during live chaos drills.

### 1.2 Core Architectural Axioms
1. **Single-Cluster Master & Agent Colocation**: The ChaosProcessor control plane (master services) and its distributed agent fleet are guaranteed to be colocated within the same internal orchestration cluster. No redundant cluster selection or cluster metadata is exposed in the primary interface.
2. **Cross-Cluster Target Isolation**: Targets exist outside the agent cluster boundary. Consequently, server-side infrastructure metrics cannot be directly scraped by the control plane.
3. **Agent-Perspective Telemetry Ground Truth**: All telemetry, latency envelopes, error rates, and failure distributions are captured strictly from the client-side perspective of the synthetic outbound load agents.
4. **Resilience & Fault Injection**: In addition to pure throughput generation, agents actively emulate transport-layer and network faults (packet jitter, artificial tail latency, TCP resets) against target sockets.

---

## 2. Information Architecture & Navigation

The application uses an aerospace-grade **persistent master frame** designed for large-format desktop screens (minimum width: 1440px).

### 2.1 Global Frame Layout
```
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ [⚡ CHAOSPROCESSOR]  Professionally breaking things before your users do...  ● GRID: NOMINAL  │
├─────────┬────────────────────────────────────────────────────────────────────────────────────┤
│         │ Top Sub-Bar: [LATENCY SLA: 99.95%] [ORBIT_POS: GEO-STATIONARY] [REFRESH] [PROFILE] │
│ Nav     ├────────────────────────────────────────────────────────────────────────────────────┤
│ Rail    │                                                                                    │
│ (Icon + │ Screen Content Area                                                                │
│ Label)  │ (Agents | Targets | Use Cases | Tests | Test Metrics Drill-down | Users)          │
│         │                                                                                    │
│ [Logout]│                                                                                    │
└─────────┴────────────────────────────────────────────────────────────────────────────────────┘
```

1. **Global Permanent Topmost Bar**:
   - Full viewport-width anchor across the entire screen (above both sidebar and content).
   - Brand glyph + bold title `CHAOSPROCESSOR`.
   - Developer tagline: *"Professionally breaking things before your users do — Orchestrating high-entropy failure so production never has to panic."*
   - Global system status pills: `GRID: NOMINAL`, `CONTROLLED ENTROPY`.
2. **Secondary Operational Sub-Bar**:
   - Live Latency SLA indicator (`99.95%`), telemetry node status, manual `Refresh`, `Clear all` triggers, and operator profile.
3. **Persistent Left Navigation Rail**:
   - Icon + permanent label below each destination.
   - Destinations in exact order:
     1. **SYS-CM** (`lightning_bolt` / `sys_cm`) — Executive Telemetry
     2. **Agents** (`dns`) — Registered Load Generator Fleet
     3. **Targets** (`my_location`) — Target Service Registry
     4. **Use Cases** (`list_alt`) — Scenario Definition Catalogue
     5. **Tests** (`play_circle`) — Test Execution Deck & Live Dispatch
     6. **Users** (`people`) — Operator & Access Management
   - Trailing anchor at base: `Logout` (`logout`).
   - Vertical divider (1px translucent cyan border) separating navigation from content.

---

## 3. Visual Language & Design System ("Aether Flight Deck")

| Token / Layer | Value / Specification |
|---|---|
| **Design System** | `Aether Flight Deck` (Deep Space Cyan) |
| **Color Mode** | Dark Aerospace Console (`#0a0e18` base canvas, `#0f131d` card surface) |
| **Primary Accent** | Quantum Cyan (`#06b6d4`, `#22d3ee`, `#00f3ff`) |
| **Success / Nominal** | Emerald Phosphor (`#10b981`, `#00ff88`) |
| **Warning / Jitter** | Amber Solar Flare (`#f59e0b`) |
| **Critical / Abort** | Crimson Nova (`#ef4444`, `#dc2626`) |
| **Typography** | `Space Grotesk` (headings, titles, telemetry displays) & `Space Mono` / `JetBrains Mono` (numerical telemetry, IP/port data, IDs) |
| **Surfaces & Borders** | Translucent cyan micro-borders (`rgba(6, 182, 212, 0.2)`), subtle radial glow, glassmorphic headers |
| **Density & Targets** | High-density administrative layout; all interactive icon buttons carry explicit tooltips and minimum 40px hit targets |

---

## 4. Screen-by-Screen Functional Specifications

### 4.1 Agents (`/agents`) — Post-Login Default Landing
- **Purpose**: Real-time management and health inspection of the distributed worker container fleet.
- **Top App Bar Actions**: `Refresh` (manual sync), `Clear all` (opens `Clear all agents` confirmation dialog).
- **Fleet KPI Summary Deck** (4 Cards):
  1. *Total Synthetic Concurrency* (e.g. `420,000 req/s`)
  2. *gRPC Fleet Ping* (e.g. `1.82 ms avg`)
  3. *Prometheus Scrapes* (e.g. `100% healthy`)
  4. *Chaos Status* (`Idle / Injected / Active`)
- **Agent Data Table**:
  - Columns: `Agent ID`, `Host / Ingress`, `gRPC Port`, `Metrics Port`, `Enabled` (`check_circle` / `cancel_outlined`), `Status` (`online`, `degraded`, `standby`), `Actions` (`edit`, `delete`).
- **Interactive Modals**:
  - *Edit Agent Dialog*: Host, gRPC Port, Metrics Port, Status, Enabled switch toggle.
  - *Delete Agent Confirmation*: *"Remove agent `<host>:<port>`?"*
  - *Clear All Confirmation*: *"This removes every agent from the database."*

### 4.2 Targets (`/targets`)
- **Purpose**: Inventory of upstream microservices, payment gateways, and API clusters scheduled for stress/chaos orchestration.
- **Top App Bar Actions**: `Refresh`.
- **Global Action**: Standard Material Floating Action Button (FAB) at bottom-right (`+` Add Target).
- **Target Summary Cards**: Health Status (`100% OK`), Protocol Split (`HTTPS`, `HTTP`, `gRPC`), Connection Reuse Ratio (`4/5 Active`), SNI Strictness.
- **Target Data Table**:
  - Columns: `Name`, `Address`, `Protocol` (color-coded badge), `Connection Reuse` (icon toggle), `SNI Validation`, `Actions` (`edit`, `delete`).
- **Interactive Modals**:
  - *Add / Edit Target Dialog*: Target Name, Target Address URL, Protocol selector (`http`, `https`, `grpc`), SNI override, Connection Reuse switch tile.
  - *Delete Target Confirmation*: *"Remove target `<name>`?"*

### 4.3 Use Cases (`/use-cases`)
- **Purpose**: Read-only catalogue of versioned synthetic traffic scenarios and payload profiles.
- **Top App Bar Actions**: `Refresh`.
- **Catalogue Structure**:
  - Vertical stack of three-line elevated cards.
  - **Leading Element**: Color-coded HTTP Method Chip (`GET` Cyan, `POST` Emerald, `PUT` Amber, `DELETE` Crimson).
  - **Title Line**: `"<name>  (<id>)"` (e.g., `Checkout Submit  (uc-checkout-p99)`).
  - **Subtitle / Details**:
    - Line 1: Monospace endpoint path (e.g., `/api/v2/checkout/confirm`).
    - Line 2: Scenario description, traffic weights, header requirements, and payload specs.

### 4.4 Tests (`/tests`) — Operational Dispatch Console
- **Purpose**: Immediate load test dispatch and live fleet execution management.
- **Top App Bar Actions**: `Refresh`.
- **Section A — "Start a Test" Form**:
  - Horizontal wrapping controls layout with instant validation:
    1. *Use Case* Dropdown (width: 260px) — `"<name> (<id>)"`
    2. *Target* Dropdown (width: 240px) — Target name
    3. *Users per Agent* Input (width: 160px) — Numeric, defaults to `1`
    4. *Agents* Input (width: 240px) — Text, defaults to `all` (accepts comma-separated UUIDs)
    5. *Start Test* Button (Filled, Quantum Cyan, `play_arrow` icon) — Transitions to *"Starting..."* spinner during dispatch.
  - Validation rule: Missing use case/target or non-positive users triggers transient snackbar *"Select a use case, a target and a positive number of users"*.
- **Section B — "Running Tests" Table**:
  - **Polling Contract**: Automatic asynchronous polling every 3 seconds without blocking UI reloads.
  - Columns: `Execution ID` (clickable link leading to metrics drilldown), `Use Case`, `Target`, `Agents` count, `Users/Agent`, `Started` timestamp, `Stop` Action Button.
  - Empty State: Centered developer message *"No tests are running"*.

### 4.5 Test Metrics Drilldown (`/tests/metrics/:id`) — Agent Perspective
- **Purpose**: Deep-dive telemetry console visualizing live test runs purely from the outbound synthetic agent perspective across cross-cluster boundaries.
- **Top Breadcrumb & Control Bar**:
  - Execution ID, Target URL, Active Run Timer (e.g. `00:14:32`), `Auto-Scrape: 3s`, `Export Dump`, and emergency `Abort Run` trigger.
  - Perspective Banner: Explaining that measurements originate purely from outbound synthetic traffic probes.
- **Top KPI Envelope**:
  - *Client Generated Load*: Real-time egress rate (req/s) + trend sparkline.
  - *Agent Round-Trip Latency*: p50, p90, p95, and p99 millisecond latency envelope against SLA ceiling.
  - *Client-Observed Failures*: Error percentage demuxing HTTP 502/504 errors, TCP connection resets, and drop ratios.
  - *Distributed Node Fleet*: Active worker pods (e.g., `8 / 8 Online`), CPU, RAM, and ping telemetry.
- **Telemetry Visualizations**:
  - *Client Request Rate & Status Demux*: Dual-axis time-series chart (200 OK vs 4xx vs 5xx vs Timeouts).
  - *Round-Trip Latency Envelope Chart*: Real-time curve mapped against the strict SLA boundary line.
  - *Active Chaos Perturbations Panel*: Real-time socket fault injection monitor (`LATENCY_TAIL_INJECTION`, `SOCKET_JITTER`, `TCP_RST_EMULATION`) with manual fault trigger.
  - *Prometheus Scraper Health*: Ingest lag, memory drift, and dropped packet counters.
- **Agent Fleet Runtime & Worker Performance Table**:
  - Detailed node-by-node telemetry (Region host, Virtual Users, Egress RPS, Observed P99, Node CPU load bar, Node Memory, Prometheus status).

### 4.6 Users (`/users`)
- **Purpose**: Operator and access key directory.
- **Columns**: `Username`, `Created` timestamp, `Actions` (`delete` icon button).
- **Interactive Modals**: *Delete User Dialog*: *"Remove user `<username>`?"*

### 4.7 Auth Screens (`/login`, `/register`)
- **Login (`/login`)**:
  - Centered minimalist column (max-width: 400px), no top app bar, no navigation rail.
  - Title: `ChaosProcessor` (`headlineMedium`).
  - Fields: `Username`, `Password`.
  - Actions: Primary filled button `Login` (with 18×18 spinner when busy), secondary text button `Register`.
- **Register (`/register`)**:
  - Centered column with top app bar ("Register").
  - Fields: `Username` (≥3 chars), `Password` (≥8 chars), `Confirm password`, `Registration secret`.
  - Actions: Primary filled button `Register`, secondary text button `Back to login`.

---

## 5. System Interactivity, Polling & Failure Contracts

1. **Non-Blocking 3-Second Polling**: The running tests matrix and agent metrics view poll active state every 3,000ms. Updates must mutate DOM in-place without triggering full table rerenders or screen flickers.
2. **Destructive Action Safeguard**: Destructive actions (`Delete`, `Clear all`, `Abort Run`) require explicit confirmation modals with verb-specific filled buttons.
3. **Transient Notification System**: Mutation confirmations and network error boundaries trigger bottom-center anchored Snackbars.
4. **Data Isolation**: Upstream target internal telemetry is explicitly isolated; all latency, throughput, and error graphs represent client socket telemetry.

---

## 6. Implementation Traceability Matrix

| Feature / Screen | Route | Key Data Model Contract | Primary Visual Pattern |
|---|---|---|---|
| **Agents** | `/agents` | `id`, `host`, `port`, `metricsPort`, `enabled`, `status` | 4-Card KPI + High-density Data Table |
| **Targets** | `/targets` | `id`, `name`, `address`, `protocol`, `connectionReuseEnabled`, `sni` | KPI Banner + Data Table + FAB |
| **Use Cases** | `/use-cases` | `id`, `name`, `method`, `path`, `description` | 3-Line Cards with Method Chips |
| **Tests Console** | `/tests` | `testExecutionId`, `useCaseId`, `targetId`, `numberOfAgents`, `simulatedUsersPerAgent`, `startTime` | Horizontal Dispatch Form + 3s Polling Table |
| **Agent Metrics** | `/tests/metrics/:id` | `egressRps`, `p50`, `p90`, `p99`, `errorDemux`, `chaosInjections`, `workerTelemetry` | Real-time Charts + Perturbation Monitor + Worker Grid |
| **Users** | `/users` | `id`, `username`, `createdAt` | Administrative Data Table |
| **Authentication** | `/login`, `/register` | `username`, `password`, `secret` | Sparse Centered Form (400px) |
