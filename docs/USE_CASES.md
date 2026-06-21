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
# Mint an API key (one-time setup)
./bin/obscura --mint-role admin

# Use the key in CI
curl -s -X POST http://localhost:8080/api/v1/scan \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"target": "staging.example.com", "profile": "quick"}'
```

## 8. Fully offline / air-gapped recon

static binary, Obscura Scan runs in restricted environments where only DNS and
direct target access are available — drop one file on a host and go.

> **Note:** "Keyless" means no API key is required, but some keyless modules still need outbound internet access (e.g., `subdomain_scan` queries crt.sh, `wayback_urls` queries the Wayback Machine). Modules that work fully offline: `dns_records`, `dns_zone_transfer`, `tls`, `tls_ciphers`, `ssl_chain`, `sec_headers`, `http_methods`, `cors`, `cookie_audit`, `robots_txt`, `tech`, `waf_detect`, `jarm_fingerprint`.
