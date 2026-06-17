# REST API

Obscura Scan exposes a small REST API plus browser/export endpoints. By default the
API is open; you can require Bearer-token auth, per-IP rate limiting, and audit
logging via configuration.

Base URL: `http://<host>:<port>` (default `http://127.0.0.1:8080`).

## Authentication

The `/api/v1/*` surface is guarded by middleware that is **off by default** and
toggled per control:

| Control | Env var | Default |
|---------|---------|---------|
| Bearer auth | `API_AUTH_ENABLED` | `false` |
| Per-IP rate limit | `API_RATE_LIMIT_ENABLED` (`API_RATE_LIMIT` req/min) | `true` (100) |
| Audit log | `AUDIT_LOG_ENABLED` | `true` |

`/healthz`, `/metrics`, and the browser/export routes are **not** token-auth'd.

### Minting an API key

```bash
./obscura --mint-key ci-bot --mint-role admin
# API key 'ci-bot' (admin) created. Store it now — it is not recoverable:
#
#   obx_1a2b3c...   <-- shown once
```

Keys are stored as a SHA-256 hash. Send the plaintext as a Bearer token (or
`X-API-Key`):

```bash
curl -H "Authorization: Bearer obx_1a2b3c..." http://127.0.0.1:8080/api/v1/ai/status
```

## Endpoints

### Health & metrics

```
GET /healthz        -> {"status":"ok","version":"9.0.0","modules":43,"active_scans":0}
GET /metrics        -> Prometheus text exposition
```

### Tasks (scan status)

```
GET /api/v1/tasks/{taskID}
```

```json
{
  "id": "…", "url": "https://example.com",
  "state": "PROGRESS",                 // PENDING | PROGRESS | SUCCESS | FAILURE
  "completed_modules": ["dns_records","tls"],
  "error": ""
}
```

### AI copilot

```
GET  /api/v1/ai/status              -> active provider + availability
POST /api/v1/ai/chat                -> {"messages":[{"role":"user","content":"…"}],"scan_id":1}
GET  /api/v1/ai/analyze/{scanID}    -> analysis of a scan (AI or rule-based fallback)
```

### Live progress (SSE)

```
GET /stream/{taskID}     -> text/event-stream; "progress" events, then a "done" event
```

### Exports

```
GET /export/json/{scanID}
GET /export/csv/{scanID}
GET /export/stix/{scanID}
GET /export/splunk-cim/{scanID}
GET /export/qradar-leef/{scanID}
GET /export/elastic-ecs/{scanID}
GET /export/pdf/{scanID}
GET /export/docx/{scanID}
```

## Launching a scan

Scans are launched through the browser form (CSRF-protected). A typical scripted
flow grabs a CSRF token from the cookie, posts the form, then polls the task:

```bash
OBX=http://127.0.0.1:8080
# 1) get a CSRF cookie + token
curl -s -c cj.txt $OBX/ >/dev/null
TOKEN=$(awk '/obx_csrf/{print $7}' cj.txt)

# 2) start a scan (modules repeated for each selected module)
RESP=$(curl -s -b cj.txt -X POST $OBX/scan \
  --data-urlencode "csrf_token=$TOKEN" \
  --data-urlencode "target=example.com" \
  --data-urlencode "mode=defensive" \
  --data-urlencode "modules=dns_records" \
  --data-urlencode "modules=tls" \
  --data-urlencode "modules=sec_headers")

# 3) the response page references /stream/{taskID}; poll the task API:
TASK=$(echo "$RESP" | grep -oE '/stream/[a-f0-9-]+' | head -1 | sed 's#/stream/##')
until curl -s $OBX/api/v1/tasks/$TASK | grep -q '"state":"SUCCESS"'; do sleep 1; done

# 4) the scan id lands in the results; fetch the JSON export
curl -s $OBX/export/json/1 | jq '._summary'
```

## Response shapes

A scan's JSON export contains one key per module (the module's `data` map) plus
`_meta` and `_summary`:

```json
{
  "_meta": { "target": "example.com", "url": "https://example.com",
             "modules": ["dns_records","tls"], "module_status": {"tls":"success"},
             "scan_id": 1, "scan_date": "…" },
  "_summary": { "risk_score": 62, "risk_level": "high",
                "total_findings": 6, "critical": 0, "high": 2, "medium": 2, "low": 2 },
  "dns_records": { "A": ["…"], "MX": ["…"] },
  "tls": { "subject": {…}, "issuer": {…}, "not_after": "…" }
}
```

## Errors

JSON endpoints return standard HTTP status codes with a JSON body
(`{"error":"…"}`). Auth failures return `401`; rate-limit exhaustion returns `429`.
