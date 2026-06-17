# Configuration

Obscura Scan is configured entirely through **environment variables** and/or a
`.env` file. **Every value is optional** — it runs with zero configuration, and
key-gated modules simply skip when their key is absent.

## Precedence & load order

1. **Real OS environment variables** (highest priority — never overwritten).
2. `.env` in the current working directory.
3. `.env` next to the binary (`os.Executable()` dir) — fallback.

> Comments in `.env` must be on their **own line**. A blank value followed by an
> inline `# comment` is parsed as the value by dotenv loaders.

Copy the template to get started:

```bash
cp .env.example .env
```

## Aliases

App-level variables use the new `OBSCURA_` prefix, with legacy `FLASK_`/`AEGIS_`
names accepted as aliases. Several providers accept multiple key spellings. The
first non-empty value wins.

| Logical key | Accepted env vars |
|-------------|-------------------|
| VirusTotal | `VT_API_KEY`, `VIRUSTOTAL_API_KEY`, `VTOTAL_KEY` |
| Shodan | `SHODAN_API_KEY`, `SHODAN_KEY` |
| GitHub | `GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_API_KEY` |
| Gemini | `GEMINI_API_KEY`, `GOOGLE_API_KEY`, `GOOGLE_GENAI_API_KEY` |
| OpenAI | `OPENAI_API_KEY`, `OPENAI_KEY` |
| Anthropic | `ANTHROPIC_API_KEY`, `CLAUDE_API_KEY` |

## Application

| Variable | Alias | Default | Description |
|----------|-------|---------|-------------|
| `OBSCURA_HOST` | `FLASK_HOST` | `127.0.0.1` | Listen address |
| `OBSCURA_PORT` | `FLASK_PORT` | `8080` | Listen port |
| `OBSCURA_DB_PATH` | `AEGIS_DB_PATH` | `obscura.db` | SQLite path (WAL) |
| `OBSCURA_CACHE_TTL` | `AEGIS_CACHE_TTL` | `3600` | Result cache TTL (seconds; `0` disables) |
| `OBSCURA_SECRET_KEY` | `FLASK_SECRET_KEY` | random | Session secret (random 32 bytes if empty) |
| `OBSCURA_DEBUG` | `FLASK_DEBUG` | `0` | Debug flag |
| `OBSCURA_ALLOW_INTERNAL` | — | `false` | Permit private/loopback targets (also `--allow-internal`) |

## Thresholds & automation

| Variable | Default | Description |
|----------|---------|-------------|
| `DEFAULT_TIMEOUT` | `15` | Per-request timeout (seconds) |
| `ALERT_THRESHOLD` | `60` | Risk score at/above which a scan sends an alert |
| `AUTO_TICKET_THRESHOLD` | `70` | Reserved |
| `MAX_CONCURRENT_SCANS` | `5` | Campaign/multi-target concurrency bound |
| `WORKFLOW_MAX_STEPS` | `15` | Reserved |
| `API_RATE_LIMIT` | `100` | REST API requests/min per IP |

## AI providers

| Variable | Default | Description |
|----------|---------|-------------|
| `AI_ENABLED` | `true` | Master switch for AI features |
| `AI_PRIMARY_PROVIDER` | `gemini` | `gemini` \| `openai` \| `anthropic` |
| `GEMINI_API_KEY` / `GEMINI_MODEL` | — / `gemini-2.5-flash` | Google Gemini |
| `OPENAI_API_KEY` / `OPENAI_MODEL` | — / `gpt-4-turbo-preview` | OpenAI |
| `ANTHROPIC_API_KEY` / `ANTHROPIC_MODEL` | — / `claude-3-sonnet-20240229` | Anthropic |

With no provider key set, AI falls back to a built-in rule-based analyst (offline).

## Threat-intel API keys (all optional)

Modules that require a key **skip gracefully** when it's unset.

| Provider | Env var | Module |
|----------|---------|--------|
| VirusTotal | `VT_API_KEY` | `virustotal` |
| Shodan | `SHODAN_API_KEY` | `shodan`, `favicon_pivot` |
| AbuseIPDB | `ABUSEIPDB_API_KEY` | `abuseipdb` |
| GreyNoise | `GREYNOISE_API_KEY` | `greynoise` |
| AlienVault OTX | `OTX_API_KEY` | `otx` |
| SecurityTrails | `SECURITYTRAILS_API_KEY` | `securitytrails` |
| URLScan | `URLSCAN_API_KEY` | `urlscan` (search works without) |
| GitHub | `GITHUB_TOKEN` | (intel) |

Extended OSINT keys are also recognized: `HUNTER_API_KEY`, `CENSYS_API_ID`,
`CENSYS_API_SECRET`, `LEAKCHECK_API_KEY`, `FOFA_EMAIL`/`FOFA_API_KEY`,
`DEHASHED_EMAIL`/`DEHASHED_API_KEY`, `FULLHUNT_API_KEY`, `ZOOMEYE_API_KEY`,
`BINARYEDGE_API_KEY`, `INTELX_API_KEY`, `BUILTWITH_API_KEY`, `WHOISXML_API_KEY`,
`HIBP_API_KEY`.

## Notifications

| Variable | Channel |
|----------|---------|
| `SLACK_WEBHOOK_URL` | Slack |
| `DISCORD_WEBHOOK_URL` | Discord |
| `TEAMS_WEBHOOK_URL` | Microsoft Teams |
| `TELEGRAM_BOT_TOKEN` + `TELEGRAM_CHAT_ID` | Telegram |

## SIEM

| Variable | Use |
|----------|-----|
| `SPLUNK_HEC_URL` / `SPLUNK_HEC_TOKEN` | Splunk HEC |
| `ELASTIC_URL` / `ELASTIC_API_KEY` | Elastic |

## Enterprise / API hardening

| Variable | Default | Description |
|----------|---------|-------------|
| `API_AUTH_ENABLED` | `false` | Require Bearer token on `/api/v1/*` |
| `API_RATE_LIMIT_ENABLED` | `true` | Per-IP rate limiting on the API |
| `AUDIT_LOG_ENABLED` | `true` | Audit-log API requests |

## Secret masking

Secrets are never logged in full. In the Settings UI and JSON export they are masked:
`""` → empty, ≤8 chars → `••••••••`, otherwise first-4 + bullets + last-4.
