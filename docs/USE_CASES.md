# Use Cases

How real teams use Obscura Scan. Every workflow below works with **zero API keys**
unless noted.

## 1. Bug bounty / pentest recon

Map a target's external attack surface fast, then drill in.

- **Profile:** *Bug Bounty*, mode *Semi-offensive*.
- **What runs:** `subdomain_scan` + `dns_bruteforce` + `subdomain_permutation`
  (discover hosts) → `subdomain_takeover` (claimable dangling CNAMEs) →
  `js_endpoints` + `source_maps` + `js_secrets` (hidden endpoints, leaked source &
  keys) → `http_probe`, `cors`, `http_methods`, `cloud_buckets`, `port_scan`.
- **Why it's strong:** the JS modules routinely surface dozens of undocumented API
  endpoints; `source_maps` can hand you the entire original source; `subdomain_takeover`
  finds the high-payoff dangling records — all without a single API key.

> Use the **attack-surface graph** to see how subdomains, IPs and ASNs relate, and
> **export to JSON/CSV** to feed your own tooling.

## 2. Continuous attack-surface monitoring (ASM)

Watch an asset over time and get alerted on change.

- Add the target on **Scheduled** with an interval (e.g. every 1440 minutes).
- Configure a notification channel (`SLACK_WEBHOOK_URL`, etc.).
- Each scheduled run is **diffed against the previous scan**; any **new finding**
  (a new subdomain, a new exposure, a new open port) fires a **change alert**.
- Review trends via **Compare** between any two scans.

This turns Obscura Scan into a lightweight, self-hosted "watchtower" for your
external footprint.

## 3. Compliance & hardening checks

Quantify and report on posture for audits (PCI-DSS, SOC 2, ISO 27001…).

- **Profile:** *Compliance*.
- **What runs:** `sec_headers`, `ssl_chain`, `tls`, `tls_ciphers`, `spf_analyzer`,
  `email_security`, `cookie_audit`, `http_methods`, `security_txt`.
- **Output:** a 0–100 risk score, severity-tagged findings, and a branded **PDF/DOCX
  report** to hand to stakeholders. The SIEM exporters (Splunk-CIM, QRadar-LEEF,
  Elastic-ECS) feed your compliance pipeline.

## 4. Threat hunting & infrastructure pivoting

Fingerprint and correlate infrastructure.

- `jarm_fingerprint` produces an active TLS server fingerprint you can pivot on
  (search Shodan/Censys for `ssl.jarm:<hash>` to find sibling/related hosts).
- `favicon_pivot` (with a Shodan key) finds every host sharing the same favicon —
  shadow infra, staging, related orgs.
- `asn_lookup` + `reverse_ip` + `cert_transparency` map the surrounding network and
  certificate footprint.
- `urlscan` / `wayback_urls` add historical context.

## 5. Brand protection & anti-phishing

Find look-alike domains before attackers use them.

- `typosquat` algorithmically generates homoglyph/typo permutations of your domain
  and **resolves them**, flagging registered look-alikes — pure DNS, no key.
- Schedule it for ongoing brand monitoring; new registrations trigger change alerts.

## 6. SIEM / SOAR integration

Feed findings straight into your security stack.

- **Exports:** STIX 2.1 (TIP ingestion), Splunk-CIM (ndjson), QRadar-LEEF,
  Elastic-ECS (ndjson) — all from `/export/{format}/{id}`.
- **REST API** (see [API](API.md)) to launch scans and pull results from automation.
- **Prometheus `/metrics`** for scan counts, module outcomes, and durations.

## 7. CI/CD security gate

Run a scan in your pipeline and gate on risk.

```bash
# pseudo-pipeline step
SCAN=$(curl -s -b cookies -X POST $OBX/scan --data-urlencode csrf_token=$T \
       --data-urlencode target=$DEPLOY_HOST --data-urlencode modules=sec_headers \
       --data-urlencode modules=tls_ciphers --data-urlencode modules=http_probe)
# ... poll /api/v1/tasks/{id}, fetch /export/json/{scan_id}, fail the build if
#     _summary.risk_score exceeds your threshold.
```

## 8. Fully offline / air-gapped recon

Because 36 modules need no external services and everything is embedded in one
static binary, Obscura Scan runs in restricted environments where only DNS and
direct target access are available — drop one file on a host and go.
