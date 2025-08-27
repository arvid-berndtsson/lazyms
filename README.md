# lazyms — Azure Security TUI (Go)

A fast, keyboard-first terminal UI for **Azure/Microsoft 365 security** work:
- Browse & query Azure resources (via **Azure Resource Graph**)
- Look up and triage **security incidents**
- View & **modify** Microsoft security policies (Conditional Access, Intune)
- Create **application update rings** (Windows Update for Business)
- Works on macOS, Linux, and Windows — mouse & keyboard

> Status: alpha. Expect dragons 🐉

## ✨ Features (alpha scope)

- LazyGit-style panes, Tab/Shift+Tab focus switching
- Mouse focus & wheel scrolling
- Preset KQL queries + custom editor
- Incidents list with detail view and quick actions
- Policy view with guarded edit/confirm flows
- Update ring creation wizard

## 🔧 Installation

### From a release (recommended)

Grab the latest asset from Releases and put `lazyms` on your PATH.

### Build from source

```bash
git clone https://github.com/arvid-berndtsson/lazyms
cd lazyms
go mod tidy
go build -o bin/lazyms ./cmd/lazyms
```

🔐 Authentication

The tool supports:
  
  1. Azure CLI login (SSO): if you’ve run az login, we’ll use it automatically.
  2. Device Code flow: if CLI is unavailable, we print a code & URL to authenticate.

You may set these (optional) in ~/.config/lazyms/config.yaml:
```yaml
tenantId: "<your-tenant-guid-or-domain>"
clientId: "<your-aad-app-client-id>"
auth: "cli"   # or "devicecode"
```
> Note: Microsoft Graph and Azure Resource Manager are different audiences; we request tokens for both as needed. Ensure your AAD app has the required permissions for the endpoints you’ll use.

## ⌨️ Keybindings

- Tab / Shift+Tab — Next / Previous pane
- Ctrl+h / Ctrl+l / ← / → — Focus left/right
- Up/Down / PgUp/PgDn — Navigate in pane
- / — Edit query / filter (left pane)
- r — Refresh
- e — Edit (policy/selection)
- o — Open in browser (incident/policy)
- Enter — View details
- q — Quit

Mouse:
- Click focuses a pane
- Wheel scrolls focused pane

## 🧪 Development

Lint & test:
```bash
golangci-lint run
go test ./...
```

Run in dev:
```bash
go run ./cmd/lazyms
```

## 🚢 Release

We use GoReleaser via GitHub Actions. To cut a release:
```bash
git tag v0.1.0
git push origin v0.1.0
```

This builds macOS (amd64/arm64), Linux (amd64/arm64), Windows (amd64), creates archives and checksums, and publishes a GitHub Release.

## ⚠️ Permissions & Safety

- Least-privilege: grant only the Graph/ARM permissions you need.
- Editing policies is powerful — every mutation shows a dry-run preview; you must confirm before applying.
- No tokens are written to disk by default.

## 📍 Roadmap

See the repository Issues for milestones:

- ARG queries + presets
- Security Incidents triage
- Conditional Access + Intune edit flows
- Update rings wizard
- Config & logging polish

## 📝 License

MIT (see LICENSE)