# 🧭 TODO (practical, ordered)

## Milestone 0 — Skeleton & UX

- Initialize repo (files above), CI, and GoReleaser.
- Add Bubble Tea scaffold with two panes (Resources | Incidents/Details), Tab/Shift+Tab, Ctrl+h/l, arrow keys, mouse focus & wheel.
- Add bubbles/table on left; viewport/markdown details on right.
- Add keymaps + help bar (bubbles/help).

## Milestone 1 — Auth & Config

- Config file ~/.config/azsec-tui/config.yaml (tenantId, clientId, preferred auth: cli|devicecode).
- Implement auth: try Azure CLI credential first; fallback to Device Code. Show “Logged in as {UPN} / {tenant}”.
- Secure token handling in memory; no on-disk tokens.

## Milestone 2 — Resource Browsing / Query

- Integrate Azure Resource Graph client (Go Azure SDK) for KQL search.
- Preset queries (All VMs, Unencrypted Storage, Public IPs, etc.); / to edit query.
- Table paging; r to refresh; status bar with counts & duration.

## Milestone 3 — Incidents & Alerts

- Microsoft Graph Security: list incidents (filter by status/severity/assignedTo).
- Right pane shows incident details; open in browser (o).
- Actions: assign, close, add comment (if permissions allow).

## Milestone 4 — Policy Modification

- Conditional Access policies: list, view JSON, enable/disable, clone (confirm modal).
- Intune device compliance/config profiles: view, edit specific fields (safe forms).
- Guardrails: dry-run preview & diff before PATCH.

## Milestone 5 — Application “Update Rings”

- Intune Windows Update for Business “update rings”: list/create/update (name, channels, deferrals).
- Wizard-like modal to create a ring; validation & confirm.

## Milestone 6 — Polish & Ops

- Logging panel (toggle with ~) + file logs on demand.
- Error surfaces with actionable messages (403, 404, throttling).
- Telemetry opt-in (env/config), never default on.
- Unit tests for parsing/render; integration tests with AZURE_TENANT_ID in CI matrix (optional).
