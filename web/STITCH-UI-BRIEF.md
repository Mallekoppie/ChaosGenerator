# Chaos Master — Web Control Panel UI Brief

> **Purpose:** This document is a complete, implementation-agnostic specification of the
> **Chaos Master** web control panel. It is written so that an AI UI design tool
> (e.g. <https://stitch.withgoogle.com>) can regenerate the interface, and so that the
> resulting design can be handed back to the codebase and re-implemented without losing
> behaviour.
>
> Keep the **labels, screen names, table columns, routes and actions** exactly as written —
> they map 1:1 to the running application.

---

## 1. Product summary

**Chaos Master** is a control panel for a distributed load/chaos-testing platform.

The user configures **test targets** (services to hammer), the platform runs **use cases**
(predefined HTTP/gRPC request scenarios) against those targets **through distriuted agents**,
and the user starts/stops **tests** and watches them run. Users also manage the registered
**agents** and platform **users**.

The app is a **data-dense administrative dashboard**, not a marketing site. Priorities:

1. Legibility of tables and forms.
2. Fast scanning of status (enabled/disabled, running/stopped).
3. Obvious primary actions.
4. Calm, professional, developer-tool aesthetic.

**Audience:** SRE / performance-engineer operators who know the domain. Favour density and
clarity over decoration.

---

## 2. Design system

### 2.1 Foundation

| Token | Value |
| --- | --- |
| Design language | Material Design 3 (Material You), Flutter's `useMaterial3: true` |
| Seed / accent colour | **Deep Purple** (Flutter `Colors.deepPurple`) — the whole palette is generated from this single seed |
| Background / surfaces | Material 3 tonal surfaces derived from the purple seed (light theme, `ColorScheme.fromSeed`) |
| Primary action colour | M3 `colorScheme.primary` (a purple in the 600–700 range) |
| Error colour | M3 `colorScheme.error` (red) — used for inline form errors and destructive confirmation buttons |
| Text colour | M3 `colorScheme.onSurface` (near-black on light surfaces) |
| Elevation | Default M3 tonal elevation; cards, app bars and dialogs use standard M3 surface tint |
| Corner radius | M3 defaults (12 px medium, 28 px for FAB and filled buttons) |
| App title | `Chaos Master` (window/document title) |

> **Important:** The palette is *derived from one seed colour*. If you change the accent,
> pick a **single seed hue** and let the tooling generate the rest of the ramp — do not
> hard-code multiple unrelated accent colours.

### 2.2 Typography

Use the platform font stack (system UI). Scale per Material 3:

| Role | M3 style | Where used |
| --- | --- | --- |
| App bar / screen titles | `titleLarge` | Screen name in the top app bar |
| Section heading | `titleLarge` | "Running tests" on the Tests screen |
| Hero / auth heading | `headlineMedium` | "Chaos Master" on the login screen |
| Body / table cells / labels | `bodyMedium` | Table data, list text, field values |
| Field labels & helper text | `bodySmall` | Underlined/floating input labels |
| Chip text | `labelLarge` | HTTP-method chip on use cases |

### 2.3 Iconography

Material Symbols / Material Icons — outlined, standard weight. **These exact icons are used:**

| Icon | Meaning |
| --- | --- |
| `dns` | Agents |
| `my_location` | Targets |
| `list_alt` | Use Cases |
| `play_circle` | Tests |
| `people` | Users |
| `logout` | Log out |
| `refresh` | Refresh current list |
| `add` | Add (floating action button) |
| `edit` | Edit row |
| `delete` | Delete row |
| `delete_sweep` | Clear all agents |
| `check_circle` (filled, primary colour) | "true" / enabled |
| `cancel_outlined` | "false" / disabled |
| `play_arrow` | Start test |

### 2.4 Component vocabulary

- **Top app bar** — one per screen, showing the screen name; holds `Refresh` (and on Agents,
  `Clear all`).
- **Navigation rail** — persistent left side navigation with labels always visible.
- **Data table** — the primary content pattern for Agents, Targets and Users. Column headers
  are plain text; rows are separated by default M3 divider lines; row actions are icon buttons.
- **Card + list tile** — used for Use Cases.
- **Floating action button** — used only on Targets ("add target").
- **Alert dialog** — used for edit forms, add forms and destructive confirmations.
- **Filled button** — the single primary action on a screen (`Login`, `Register`, `Save`,
  `Delete` in confirm dialogs, `Start test`).
- **Text button** — secondary/cancel actions (`Cancel`, `Register`, `Back to login`, `Stop`).
- **Snack bar** — transient feedback for success and error messages, anchored bottom-centre.
- **Circular progress indicator** — full-screen centred spinner while first load is in flight;
  inline 18×18 spinner inside a busy primary button.
- **Empty state** — centred plain-text message when a list has no rows.
- **Error state** — centred plain-text error message when a list fails to load and is empty.

---

## 3. Global layout & navigation

```
┌──────────────────────────────────────────────────────────────────────┐
│  ┌──────────┐  │                                                     │
│  │ Agents   │  │   ┌────────────────────────────────────────────┐    │
│  │ Targets  │  │   │  AppBar: <Screen name>     [refresh] [...]  │    │
│  │ Use Cases│  │   ├────────────────────────────────────────────┤    │
│  │ Tests    │  │   │                                            │    │
│  │ Users    │  │   │            screen content                  │    │
│  │          │  │   │                                            │    │
│  │ [logout] │  │   │                                            │    │
│  └──────────┘  │   └────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────┘
   NavigationRail   │  Expanded content area
   (labels always   │
    shown)         VerticalDivider
```

- Root layout is a full-height **`Row`** wrapped in a `SafeArea`.
- **Left:** a `NavigationRail` with `labelType: all` (icon **and** text label under every
  destination). Destinations, top to bottom:
  1. **Agents** — `dns`
  2. **Targets** — `my_location`
  3. **Use Cases** — `list_alt`
  4. **Tests** — `play_circle`
  5. **Users** — `people`
- **Trailing (bottom of the rail):** a `logout` icon button with tooltip **"Log out"**.
- A **1 px vertical divider** separates the rail from the content.
- **Right:** the active screen, filling the remaining width.

### 3.1 Routing

| Route | Screen | Notes |
| --- | --- | --- |
| `/login` | Login | Entry point when unauthenticated |
| `/register` | Register user | Public |
| `/agents` | Agents | **Default landing screen after login** |
| `/targets` | Targets | |
| `/use-cases` | Use Cases | |
| `/tests` | Tests | |
| `/users` | Users | |

**Auth guard behaviour:**
- Unauthenticated users are redirected to `/login` from any protected route.
- Authenticated users hitting `/login` or `/register` are redirected to `/agents`.
- Logging out from the rail navigates back to `/login`.

---

## 4. Screen specifications

### 4.1 Login (`/login`)

A minimal, centred authentication screen. **No app bar, no navigation rail.**

- Background: plain theme surface.
- A single centred column, **max width 400 px**, **24 px padding**, transparent (no card
  box is drawn — it is a bare centred column).
- Contents, top to bottom:
  1. **Heading** — "Chaos Master", `headlineMedium`, centre-aligned.
  2. 24 px gap.
  3. **Username** text field, label "Username".
  4. 12 px gap.
  5. **Password** text field, label "Password", obscured (dots).
  6. *(Conditionally)* 12 px gap + **inline error message** in the error colour.
  7. 24 px gap.
  8. **Primary button** — "Login" (filled, full width). While signing in it is disabled and
     shows an 18×18 circular spinner instead of the label.
  9. 8 px gap.
  10. **Secondary text button** — "Register" (full width), navigates to `/register`. Disabled
      while busy.
- Pressing Enter in either field submits the form.

**Visual note:** this is the only "hero" moment in the app. Keep it sparse and calm;
the single accent-coloured filled button should be the focal point.

---

### 4.2 Register (`/register`)

Same centred-column treatment as Login (**max width 400 px, 24 px padding**), but wrapped in
a **Form** with validation, and it **has a top app bar**.

- **Top app bar:** title "Register". No actions.
- Contents, top to bottom (12 px gaps between fields):
  1. **Username** — label "Username"; validation: at least 3 characters.
  2. **Password** — label "Password", obscured; validation: at least 8 characters.
  3. **Confirm password** — label "Confirm password", obscured; validation: must equal
     the password field.
  4. **Registration secret** — label "Registration secret" (plain text field, not next to
     the other two); validation: required.
  5. *(Conditionally)* inline **error message** in the error colour.
  6. 24 px gap.
  7. **Primary button** — "Register" (filled, full width). Shows the 18×18 spinner while busy.
  8. 8 px gap.
  9. **Secondary text button** — "Back to login".
- Validation errors appear under each field in the error colour.

---

### 4.3 Agents (`/agents`)

The default screen after login.

**Top app bar**
- Title: **"Agents"**
- Actions (trailing, left→right):
  1. `refresh` icon button, tooltip **"Refresh"** — reloads the list.
  2. Text-button-with-icon **"Clear all"**, leading `delete_sweep` icon — opens the
     *Clear all agents* confirmation dialog.

**Content — data table** (16 px padding around it; horizontally scrollable if narrow):

| Column | Content |
| --- | --- |
| **ID** | Agent identifier (string, can be a UUID) |
| **Host** | Hostname / IP |
| **Port** | Agent gRPC port (number) |
| **Metrics** | Metrics port (number) |
| **Enabled** | `check_circle` in the accent/primary colour when true, `cancel_outlined` when false |
| **Status** | Free-text status string (e.g. online/offline) |
| **Actions** | `edit` icon button (tooltip "Edit") + `delete` icon button (tooltip "Delete") |

**States**
- *Loading & empty* → centred circular progress indicator.
- *Error & empty* → centred error text.
- *Empty & idle* → centred message **"No agents are connected"**.

**Dialogs**
- **Edit agent** (`AlertDialog`, title "Edit agent"): form with
  **Host** (text), **Port** (numeric), **Metrics port** (numeric), **Status** (text), and a
  **switch list tile** titled "Enabled". Actions: **Cancel** (text button) / **Save** (filled).
- **Delete agent** (`AlertDialog`, title "Delete agent"): message
  **"Remove agent `<host>:<port>`?"**. Actions: **Cancel** / **Delete** (filled).
- **Clear all agents** (`AlertDialog`, title "Clear all agents"): message
  **"This removes every agent from the database."** Actions: **Cancel** / **Clear all** (filled).

---

### 4.4 Targets (`/targets`)

**Top app bar**
- Title: **"Targets"**
- Action: `refresh` icon button, tooltip **"Refresh"**.

**Floating action button**
- A standard M3 `FloatingActionButton` with an `add` icon, tooltip **"Add target"**, bottom-right.

**Content — data table** (16 px padding):

| Column | Content |
| --- | --- |
| **Name** | Human-friendly target name |
| **Address** | URL/endpoint, e.g. `http://localhost:8080` |
| **Protocol** | One of `http`, `https`, `grpc` |
| **Connection reuse** | `check_circle` / `cancel_outlined` icon |
| **SNI** | Server Name Indication string (may be empty) |
| **Actions** | `edit` icon button (tooltip "Edit") + `delete` icon button (tooltip "Delete") |

**States**
- *Loading & empty* → centred spinner.
- *Error & empty* → centred error text.
- *Empty & idle* → centred message **"No targets registered"**.

**Dialogs**
- **Add / Edit target** (`AlertDialog`, title **"Add target"** or **"Edit target"**): form with
  - **Name** (text)
  - **Address** (text, hint `http://localhost:8080`)
  - **Protocol** (dropdown: `http` / `https` / `grpc`; defaults to `http`)
  - **SNI (optional)** (text)
  - **Connection reuse** — switch list tile

  Actions: **Cancel** (text) / **Save** (filled).
- **Delete target**: title "Delete target", message **"Remove target `<name>`?"**,
  actions **Cancel** / **Delete** (filled).

---

### 4.5 Use Cases (`/use-cases`)

Read-only catalogue screen.

**Top app bar**
- Title: **"Use Cases"**
- Action: `refresh` icon button, tooltip **"Refresh"**.

**Content — list of cards** (16 px page padding, cards stacked vertically):

Each row is a **Card** containing a `ListTile` (three-line):
- **Leading:** a **Chip** whose label is the HTTP method (e.g. `GET`, `POST`) — this is the
  natural place to colour-code methods (GET blue, POST green, PUT amber, DELETE red) while
  keeping the rest of the app purple.
- **Title:** `"<name>  (<id>)"` — the use case name, two spaces, then the id in parentheses.
- **Subtitle:** `"<path>\n<description>"` — the request path on line 1, description on line 2.
- `isThreeLine: true`.

**States**
- *Loading & empty* → centred spinner.
- *Error & empty* → centred error text.
- (No explicit empty-state message; an empty catalogue simply renders no cards.)
- **No add/edit/delete** — use cases are defined by the backend and are static in the UI.

---

### 4.6 Tests (`/tests`)

The operational screen: start a test, then watch running tests. Two vertical sections.

**Top app bar**
- Title: **"Tests"**
- Action: `refresh` icon button, tooltip **"Refresh"**.

**Section A — "Start a test" form** (16 px padding, laid out as a wrapping row of controls,
16 px spacing, vertically centred):

| Control | Width | Detail |
| --- | --- | --- |
| **Use case** dropdown | 260 px | Options rendered as `"<name> (<id>)"`; label "Use case" |
| **Target** dropdown | 240 px | Options rendered as target **name**; label "Target" |
| **Users per agent** text field | 160 px | Numeric; **default value `1`**; label "Users per agent" |
| **Agents** text field | 240 px | Label "Agents", hint **"all or comma separated ids"**; **default value `all`** |
| **Start test** button | auto | Filled button with a `play_arrow` icon; label becomes **"Starting…"** and disables while in flight |

- Validation: starting requires a selected use case **and** target **and** a users value > 0.
  Otherwise a snack bar shows
  **"Select a use case, a target and a positive number of users"**.
- On success a snack bar shows **"Test started"**; on failure it shows the server message.

**Section B — "Running tests"** (24 px above the heading):
- Heading: **"Running tests"** in `titleLarge`.
- **Data table** (fills remaining vertical space, scrollable):

| Column | Content |
| --- | --- |
| **Execution** | Test execution id |
| **Use case** | Use case name, falling back to id if the name is empty |
| **Target** | Target name, falling back to id if the name is empty |
| **Agents** | Number of agents participating |
| **Users/agent** | Simulated users per agent (number) |
| **Started** | Local timestamp derived from the start time (milliseconds → epoch, shown via the platform's default date-time format) |
| *(blank header)* | **Stop** text button — stops that execution |

- **Auto-refresh:** the running-tests table **polls every 3 seconds** while this screen is
  open. Design for a table whose numbers/timestamps change in place without a full-page
  reload — do not show a blocking spinner on each poll.
- **Empty state:** centred message **"No tests are running"**.

---

### 4.7 Users (`/users`)

**Top app bar**
- Title: **"Users"**
- Action: `refresh` icon button, tooltip **"Refresh"**.

**Content — data table** (16 px padding):

| Column | Content |
| --- | --- |
| **Username** | Account name |
| **Created** | Local timestamp derived from creation time (milliseconds → epoch, default date-time format) |
| **Actions** | `delete` icon button, tooltip "Delete" |

**States**
- *Loading & empty* → centred spinner.
- *Error & empty* → centred error text.

**Dialog**
- **Delete user**: title "Delete user", message **"Remove user `<username>`?"**,
  actions **Cancel** / **Delete** (filled).

---

## 5. Interaction & feedback rules

1. **Refresh** buttons reload only their own screen's data; there is no global refresh.
2. **Destructive actions always confirm** via an alert dialog with a filled confirm button
   labelled with the specific verb (`Delete`, `Clear all`) — never a generic "OK".
3. **Snack bars** report the outcome of mutations:
   - success (e.g. "Test started")
   - failure (server-provided message, or a fallback like "Failed to update agent").
4. **Busy buttons** swap their label for an 18×18 circular spinner and become disabled.
5. **Inline errors** on the auth screens and inside dialogs are plain text in the error colour.
6. **Loading contract:** a screen shows the full-screen spinner only on its *first* load when
   the list is empty. Later refreshes update in place; the table keeps showing stale rows.
7. **Accessibility / affordance:** every icon-only button must carry a tooltip
   ("Refresh", "Edit", "Delete", "Add target", "Log out"). Row actions live at the right edge
   of each table row. Keep tap targets ≥ 40 px.

---

## 6. Data model reference (drives the tables/forms)

Keep these field names — they are the contract between UI and backend.

**Agent**
`id` · `host` · `port` · `metricsPort` · `enabled` (bool) · `status`

**Target**
`id` · `name` · `address` · `protocol` (`http` | `https` | `grpc`) · `connectionReuseEnabled` (bool) · `sni`

**UseCase**
`id` · `name` · `method` · `path` · `description`

**RunningTestExecution**
`testExecutionId` · `useCaseId` · `useCaseName` · `targetId` · `targetName` · `numberOfAgents` · `simulatedUsersPerAgent` · `startTime`

**User**
`id` · `username` · `createdAt`

---

## 7. What to keep stable when redesigning (round-trip contract)

So the new design can be re-implemented against the existing backend, **do not rename or
remove**:

- The **five navigation destinations** and their order, and the **log-out** action.
- The **routes** `/login`, `/register`, `/agents`, `/targets`, `/use-cases`, `/tests`, `/users`
  and the rule that `/agents` is the post-login default.
- The **exact labels** listed for buttons, dialog titles/messages, empty states and validation
  errors (they are asserted in tests and mirrored in the code).
- The **table column sets** on Agents, Targets, Tests and Users, and the **card layout** of
  Use Cases.
- The **edit/add/delete/clear-all/start/stop** actions and where they live.
- The **3-second polling** of running tests.

**Free to change:** spacing, type scale within the M3 roles, the specific purple hue (as long
as it comes from a single seed), elevation, corner radii, the HTTP-method chip colours, and any
added visual polish (icons, emphasis, density controls) — provided all content still fits and
the contract above is intact.

---

## 8. One-paragraph prompt for Stitch

> Design a Material Design 3 admin dashboard for **"Chaos Master"**, a load/chaos-testing
> control panel, using a light theme generated from a single **deep purple** seed colour.
> Layout is a persistent left **navigation rail** (labels always visible) with five items —
> **Agents** (`dns`), **Targets** (`my_location`), **Use Cases** (`list_alt`), **Tests**
> (`play_circle`), **Users** (`people`) — and a **log-out** icon button at its base, separated
> by a 1 px vertical divider from the content area. Each screen has a top app bar with the
> screen name and a **Refresh** icon action; Agents also has a **Clear all** action, and
> Targets has an **add** floating action button. Agents, Targets and Users use dense
> **data tables** (columns: Agents = ID, Host, Port, Metrics, Enabled, Status, Actions;
> Targets = Name, Address, Protocol, Connection reuse, SNI, Actions; Users = Username,
> Created, Actions) with **edit/delete** icon buttons per row and check/cancel icons for
> booleans. Use Cases is a list of **cards** with an HTTP-method **chip** on the left and
> three-line title/subtitle. Tests shows a horizontal **start-a-test form** (use-case
> dropdown, target dropdown, "Users per agent" numeric field defaulting to 1, "Agents" field
> defaulting to "all", and a purple **Start test** button with a play icon) above a
> **"Running tests"** table (Execution, Use case, Target, Agents, Users/agent, Started, and a
> **Stop** button) that refreshes every 3 seconds. Login and Register are sparse, centred
> forms (max width 400 px) with a purple filled primary button. Use filled buttons for primary
> actions, text buttons for cancel, alert dialogs for add/edit and destructive confirmations,
> snack bars for feedback, centred spinners for loading, and simple centred text for empty
> states. Professional, calm, information-dense developer-tool aesthetic.
