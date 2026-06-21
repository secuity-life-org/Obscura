# Obscura Scan — Migration Notes (Python AEGIS → Go)

> [!NOTE]
> **Internal development log.** This document tracks the Python AEGIS → Go Obscura Scan migration. It may contain stale planning data and does not reflect the current feature set. For the current state of the project, see [README.md](../README.md) and [CHANGELOG.md](../CHANGELOG.md).

Decisions, deviations, and TODOs for the port of **AEGIS v9.0.0** (Python/Flask)
to **Obscura Scan** (Go single static binary). This file is updated per phase.

## Rebrand summary (§0)
- Product/display name: **Obscura Scan** (UI, reports, CLI banner, `--version`).
- Go module: `obscurascan` (no spaces allowed in module/import paths).
- Binary: `obscura`. Config struct: `ObscuraConfig`. Default DB: `obscura.db`.
- Source Python names are **unchanged** when read (aegis.py, AegisConfig,
  VT_API_KEY, …). Only the new Go project is branded.
- AI persona "AEGIS AI" → "Obscura Scan AI" (deferred to Phase 7; strings copied
  verbatim otherwise).

## Project root
`/home/sudo3rs/Documents/PrivateTools/Obscura Scan` — the directory name contains
a space; this is fine for the project root but never appears in import paths.

---

## Phase 0 — Scaffold, config, DB (DONE)

### Package layout
Created the idiomatic layout from §2: `cmd/obscura`, `internal/{config,store,
engine,modules,ai,ml,intel,export,notify,server,safety}`, `web/static/img`.
Later-phase packages are present as empty directories to be filled in order.

### Config (`internal/config`) — port of `core/config.py`
- `ObscuraConfig` mirrors `AegisConfig` keys/defaults exactly.
- **Improvement: alias resolution.** `getenvAny(...)` returns the first
  non-empty trimmed value; `aliasTable` is the single auditable source of
  accepted spellings (VT/Shodan/GitHub/Gemini/OpenAI/Anthropic + the OBSCURA_/
  legacy FLASK_/AEGIS_ app-var aliases).
- **Load order:** OS env (never overwritten) > `./.env` > `<binary-dir>/.env`,
  via `godotenv.Load` (which does not override already-set vars). No `.env` =
  proceed silently; all keys optional.
- `Validate()` returns **warnings, never fatal errors** (logged at WARN via
  slog). Same four warnings as the Python version (no AI key, gemini-primary-
  but-empty, no VT, no Shodan).
- `ConfiguredAPIKeys()` / `ConfiguredNotifications()` / `Mask()` / `Sanitized()`
  ported 1:1 (same masking rule: ""→""; len≤8→8 bullets; else first4+bullets+last4).
- Added `AllowInternal` (OBSCURA_ALLOW_INTERNAL) ahead of the Phase 1 SSRF guard.
- `config.Get()` is the `sync.Once` singleton equivalent of `lru_cache get_config()`.

### Store (`internal/store`) — port of aegis.py `init_db` + enterprise schema
- Driver: `modernc.org/sqlite` (pure Go → keeps `CGO_ENABLED=0`).
- WAL enabled via DSN pragmas (`journal_mode(WAL)`, `busy_timeout(5000)`,
  `foreign_keys(ON)`).
- All 10 tables reproduced verbatim: 5 core (`tasks`, `scans`,
  `scheduled_scans`, `ai_conversations`, `scan_notes`) + 5 enterprise
  (`scan_tags`, `scan_templates`, `api_keys`, `audit_log`, `bulk_campaigns`),
  plus the original indexes. Migrations are idempotent `CREATE TABLE IF NOT EXISTS`.
- **Legacy import:** on startup, if `obscura.db` is absent but `aegis.db` exists
  in the same dir, it (and its `-wal`/`-shm` sidecars) is renamed to `obscura.db`
  and the action is logged — existing scan history preserved.
- Repositories (ScanRepo, TaskRepo, …) are stubbed for Phase 1+ where the engine
  needs them.

### Entrypoint (`cmd/obscura`)
- Loads config, opens the WAL DB, logs a sanitized boot summary. `--version`
  prints build info (ldflags-injected). `--allow-internal` flips the SSRF
  override. Banner shows the "Obscura Scan" brand.

### Packaging
- `Makefile` (build, build-all cross-compile matrix, test `-race`, lint, docker).
- Multi-stage `Dockerfile` → `distroless/static` final image, `CGO_ENABLED=0`.
- `.env.example` lists every key grouped with comments + alias hints, all blank.

---

## Phase 1 — Engine core (DONE)
- `internal/safety`: SSRF guard at the dialer `Control` hook (re-validates the
  resolved IP on every dial/redirect), target intake validation/normalization
  (domain|ip|url|email), default-deny with `AllowInternal` opt-in.
- `internal/httpx`: shared client porting `core/utils.py fetch_with_retry` —
  exponential backoff + full jitter, bounded retries, per-host circuit breaker
  (open after 5 failures, half-open after 60s), redirect cap, SSRF-guarded
  transport.
- `internal/engine`: `Module` interface + `ModuleResult`, global `Register`/
  `Lookup` registry, concurrency-safe `SharedState` (RWMutex), DAG scheduler
  (`engine.Run`) with a bounded worker-pool semaphore, panic isolation per
  module, and graceful key-based skipping (`config.APIKey`).
- `internal/engine` task runner (`Runner`, ports `core/scanner.py`): result
  cache (superset module match within CacheTTL), task PROGRESS updates, scans
  persistence with `_meta.scan_id`, SUCCESS/FAILURE marking.
- `internal/store` repositories: `TaskRepo`, `ScanRepo`.

## Phase 2 — First modules (DONE)
Ported: `dns_records` (miekg/dns: A/AAAA/MX/NS/TXT/SOA), `tls`
(crypto/tls via the guarded dialer), `http_probe` (exposed-file scanner with
content validators). Remaining modules from §6 are TODO.

## Phase 5 — Web server (DONE)
- `internal/server`: chi router; embedded `html/template` UI (base + index +
  progress + results + scans + modules); embedded static via the `web` package
  (`//go:embed`). Routes: `/`, `POST /scan`, `/stream/{task}` (SSE), `/view/{id}`,
  `/view/by-task/{task}`, `/scans`, `/modules`, `/healthz`, `/export/json/{id}`,
  `/api/v1/tasks/{task}`.
- Middleware: RealIP, Recoverer, slog request logging, double-submit-cookie CSRF
  (browser POSTs enforced, `/api/v1/*` exempt).
- Graceful shutdown via `http.Server.Shutdown`; scans tie to the root context so
  SIGTERM cancels in-flight work.
- Verified end-to-end: dashboard renders, static loads, CSRF 403/200, scan of
  example.com completes, results render, `--version`, graceful drain.

### Deviations to revisit
- **Tailwind/Font-Awesome/htmx via CDN** (§17 wants them compiled+embedded). Kept
  CDN for now to reach a working UI; embedding/purging is a deferred perf task.
  **Update:** All front-end assets are now embedded via `//go:embed` — no CDN dependency in production.
- **Templates are freshly authored** (not a 1:1 port of the 9 Jinja2 files / the
  1458-line macro). They reuse `aegis.css` + the existing class names for the
  dark glass theme (`styles.css (formerly aegis.css)`). Full template parity (compare/graph/report/queue/scheduled/
  settings pages) is TODO.
- **`.env.example` format fix:** dotenv loaders (godotenv) keep an inline
  `# comment` as the value when the value is otherwise blank — so all comments
  now live on their own lines. (Caught when a copied `.env` set host to a comment.)

## Deferred / TODO (later phases)
- ✅ **Phase 1:** Module interface + registry, bounded worker pool, SharedState,
  httpx retry/circuit-breaker, and the SSRF-guarded dialer in `internal/safety`
  (the `AllowInternal` flag is already plumbed).
- ✅ **Phase 2–4:** Port every module; ML IsolationForest in `internal/ml`
  (will document if heuristic-scorer parity is chosen over exact port).
- ✅ **Phase 5:** chi router, all 62 routes, SSE, htmx fragment rendering, CSRF,
  templates via `//go:embed`.
- **Phase 6–8:** exporters, notifications, multi-provider AI + rule fallback,
  scheduler.
- **UI design-level rebrand** (pitch-black / Chakra Petch identity) is explicitly
  out of scope for the port (name + logo only); recorded here as a future idea.
