# Usage

This guide covers how to drive Obscura Scan day to day — the CLI, the web UI, and
every major workflow.

## Table of Contents

- [Command-line flags](#command-line-flags)
- [The web UI](#the-web-ui)
  - [Run a scan](#run-a-scan)
  - [Read the results](#read-the-results)
  - [Attack-surface graph](#attack-surface-graph)
  - [Compare two scans](#compare-two-scans)
  - [Notes](#notes)
- [Profiles (scan templates)](#profiles-scan-templates)
- [Scheduling & continuous monitoring](#scheduling--continuous-monitoring)
- [Campaigns (bulk scanning)](#campaigns-bulk-scanning)
- [Exporting](#exporting)
- [AI copilot](#ai-copilot)
- [REST API](#rest-api)
- [Observability](#observability)
- [Troubleshooting](#troubleshooting)

## Command-line flags

```
obscura [flags]

  --version             print build info (version, commit, date) and exit
  --allow-internal      permit scanning private/loopback/metadata targets
                        (default: deny — authorized internal use only)
  --mint-key NAME       create a REST API key named NAME, print it once, and exit
  --mint-role ROLE      role for --mint-key: viewer | admin   (default: viewer)
```

Host/port/database and everything else are configured via environment variables —
see [Configuration](CONFIGURATION.md).

```bash
./obscura                         # start the server on 127.0.0.1:8080
OBSCURA_PORT=9090 ./obscura       # custom port
./obscura --allow-internal        # scan 10.0.0.0/8, 127.0.0.1, etc. (authorized only)
./obscura --mint-key ci --mint-role admin   # create an API key and exit
```

## The web UI

Open <http://127.0.0.1:8080>. The sidebar gives you:

| Page | Purpose |
|------|---------|
| **Dashboard** | Launch a scan; pick a profile; recent scans |
| **History** | All completed scans |
| **Scheduled** | Recurring scans / continuous monitoring |
| **Campaigns** | Bulk multi-target scanning |
| **Modules** | Catalog of all 43 modules |
| **Settings** | Config, providers, API keys, audit log |

### Run a scan

1. On the **Dashboard**, enter a target — a domain (`example.com`), URL
   (`https://example.com/app`), or IP (`93.184.216.34`).
2. Choose a **Profile** (Quick / Full / Bug Bounty / Compliance) or hand-pick modules.
3. Pick a **mode**: *Defensive* (safe) or *Semi-offensive* (enables probing/port-scan/etc).
4. Click **Start Scan**. You'll see **live progress** stream in, then be redirected to
   the results.

### Read the results

The results page is a **report**, not a JSON dump:

- A **risk gauge** (0–100) + severity counts.
- A severity-sorted **Findings** list across all modules.
- The **AI Copilot** panel — click *Analyze* for a narrative, or ask questions.
- Per-module **human-readable cards** (tables, labeled fields), each with a
  collapsible **Raw JSON** fallback.
- Buttons: **Graph**, **Compare prev**, and exports.

### Attack-surface graph

Click **Graph** on any result for an interactive map of
`target → subdomains → IPs → ASNs → cert issuers`. Drag nodes, scroll to zoom.

### Compare two scans

When a target has more than one scan, a **Compare prev** button appears. The compare
view shows **new findings**, **resolved findings**, and the **risk-score delta**.

### Notes

Add analyst notes to any scan from its results page (stored locally).

## Profiles (scan templates)

Four presets are seeded on first run and shown as one-click buttons on the Dashboard:

| Profile | Focus |
|---------|-------|
| **Quick** | Fast baseline: DNS, TLS, headers, tech, exposed files |
| **Full** | Broad keyless recon across every category |
| **Bug Bounty** | Subdomain discovery, takeover, CORS, dorking, wayback |
| **Compliance** | Headers, TLS, email auth, cookies |

## Scheduling & continuous monitoring

On **Scheduled**, add a target with an **interval (minutes)**. Obscura Scan will
re-scan on that cadence. Because scheduled scans run in **monitoring mode**, each run
is **diffed against the previous scan** and any **new findings** trigger a
**change alert** to your configured notification channels — turning Obscura Scan into
a watchtower ("alert me when example.com exposes something new").

Configure channels via env vars (see [Configuration](CONFIGURATION.md)):
`SLACK_WEBHOOK_URL`, `DISCORD_WEBHOOK_URL`, `TEAMS_WEBHOOK_URL`,
`TELEGRAM_BOT_TOKEN` + `TELEGRAM_CHAT_ID`. Threshold alerts also fire when a scan's
risk score meets `ALERT_THRESHOLD`.

## Campaigns (bulk scanning)

On **Campaigns**, paste a list of targets (one per line) and launch. Each target gets
its own scan, and the campaign view aggregates per-target risk with links to each
result — fleet-scale attack-surface assessment in one place.

## Exporting

From any result, export to: **PDF**, **DOCX**, **JSON**, **CSV**, **STIX 2.1**,
**Splunk-CIM**, **QRadar-LEEF**, **Elastic-ECS**. The SIEM formats are ready for
direct ingestion; PDF/DOCX are branded reports for stakeholders.

```bash
curl -o report.pdf  http://127.0.0.1:8080/export/pdf/1
curl -o scan.json   http://127.0.0.1:8080/export/json/1
curl       http://127.0.0.1:8080/export/elastic-ecs/1
```

## AI copilot

If `AI_ENABLED=true` (default) and a provider key is set (`GEMINI_API_KEY`,
`OPENAI_API_KEY`, or `ANTHROPIC_API_KEY`), the copilot uses that model. With **no
keys**, it falls back to a built-in **rule-based** analyst that works fully offline —
it never hard-fails. Provider order honors `AI_PRIMARY_PROVIDER`.

## REST API

Drive scans and pull results programmatically — see the [API reference](API.md). The
API can be locked down with Bearer-token auth, per-IP rate limiting, and an audit
log.

## Observability

- `GET /healthz` — JSON health/build info.
- `GET /metrics` — Prometheus exposition (scan counts, module outcomes, durations,
  HTTP response classes).
- Structured logs on stderr; `SIGINT`/`SIGTERM` drains in-flight scans before exit.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `bind: address already in use` | Another process is using port 8080. Set `OBSCURA_PORT=9090` or stop the conflicting process. |
| Scan returns no results for a valid domain | Check your network connectivity. Some modules require outbound HTTPS access. |
| SSRF guard blocks a legitimate internal target | Use `--allow-internal` (only for authorized internal engagements). |
| Module shows "skipped — key not configured" | The module requires an API key. See [Configuration](CONFIGURATION.md) for the relevant env var. |
| Database locked errors | Ensure only one instance of Obscura is running against the same `obscura.db` file. |
| AI copilot returns generic responses | Configure at least one AI provider key (Gemini, OpenAI, or Anthropic). The rule-based fallback is used when no keys are set. |
