# Modules

Obscura Scan ships **43 modules**. **36 are fully keyless** (talk directly to the
target / DNS / TLS / HTTP). The 7 key-gated modules **skip gracefully** when their
key is unset — the scan always continues.

Categories: `recon` · `passive` · `semi-offensive` · `intel` · `analysis`.
Mode gating: *semi-offensive* modules run only when a scan's mode is set to
**Semi-offensive**.

## DNS & domain

| Module | Description | Key |
|--------|-------------|-----|
| `dns_records` | A/AAAA/MX/NS/TXT/SOA records | — |
| `dns_zone_transfer` | Tests AXFR against every nameserver (critical if it succeeds) | — |
| `dns_bruteforce` | Active subdomain brute-force from a **built-in wordlist**, with wildcard detection | — |
| `subdomain_scan` | Passive subdomain enumeration from certificate transparency (crt.sh) | — |
| `subdomain_permutation` | Mutates discovered subdomains to find hidden assets | — |
| `subdomain_takeover` | Dangling-CNAME takeover detection via service fingerprints | — |
| `reverse_ip` | Co-tenant domains on the same IP (shared hosting) | — |
| `whois` | Registration data (registrar, dates, nameservers) | — |
| `typosquat` | Generates typo/homoglyph look-alike domains and resolves them (dnstwist-style) | — |

## TLS & certificates

| Module | Description | Key |
|--------|-------------|-----|
| `tls` | Basic certificate info (subject, issuer, validity, protocol) | — |
| `ssl_chain` | Deep chain analysis: wildcard/CA trust, expiry, scoring/grade | — |
| `tls_ciphers` | Enumerates supported TLS versions & cipher suites; flags TLS 1.0/1.1, weak ciphers, no-PFS (testssl-lite) | — |
| `jarm_fingerprint` | Active JARM TLS server fingerprint (shared-infra / C2 hunting) | — |
| `cert_transparency` | crt.sh certificate history; wildcard/recent-issuance/CA-reuse flags | — |

## Web posture

| Module | Description | Key |
|--------|-------------|-----|
| `crawler` | Same-host crawl: links, forms, scripts, emails, API endpoints, supply-chain | — |
| `tech` | Technology fingerprint from headers + HTML markers | — |
| `waf_detect` | WAF/CDN detection from response headers | — |
| `sec_headers` | Security-header analysis (CSP, HSTS, X-Frame-Options, …) | — |
| `robots_txt` | robots.txt — disallowed paths & sitemaps | — |
| `security_txt` | RFC 9116 security.txt presence + contacts | — |
| `http_methods` | Allowed HTTP methods (OPTIONS); flags PUT/DELETE/TRACE | — |
| `cors` | CORS policy test — wildcard / reflected-Origin-with-credentials | — |
| `cookie_audit` | Set-Cookie flags — Secure / HttpOnly / SameSite | — |
| `http_probe` | Exposed-file & misconfig probing with content validation (semi-offensive) | — |
| `wayback_urls` | Internet Archive CDX mining for historical/forgotten URLs | — |

## JavaScript analysis

| Module | Description | Key |
|--------|-------------|-----|
| `js_endpoints` | Extracts hidden API paths, URLs, and parameters from linked JS (LinkFinder-style) | — |
| `js_secrets` | Scans inline + linked JS for leaked keys/tokens/private keys (redacted) | — |
| `source_maps` | Detects exposed `.js.map` source maps (original source-code leak) | — |

## Email authentication

| Module | Description | Key |
|--------|-------------|-----|
| `spf_analyzer` | SPF flattening, lookup counting, DMARC, policy grading | — |
| `email_security` | Full posture: DKIM selectors, DMARC, BIMI, MTA-STS, TLS-RPT, DNSSEC, CAA | — |

## Offensive surface

| Module | Description | Key |
|--------|-------------|-----|
| `port_scan` | TCP-connect scan of common ports (SSRF-guarded); flags risky services | — |
| `cloud_buckets` | S3/GCS/Azure bucket name permutation + HTTP probe (open vs exists) | — |

## OSINT & intel

| Module | Description | Key |
|--------|-------------|-----|
| `ip_geolocation` | Resolves IPs and enriches via ip-api.com (free, no key) | — |
| `asn_lookup` | ASN/BGP/owner via **Team Cymru DNS** (no API) | — |
| `google_dorking` | Generates dork queries + probes link-aggregator/social services | — |
| `urlscan` | URLScan.io public search (historical scans, IPs, servers) | — |
| `favicon_pivot` | Favicon MurmurHash3 + Shodan pivot to find related hosts | `SHODAN_API_KEY` |
| `virustotal` | Domain reputation (malicious/suspicious engine counts) | `VT_API_KEY` |
| `shodan` | Host exposure: open ports, banners, known vulns | `SHODAN_API_KEY` |
| `abuseipdb` | IP abuse-confidence score & report count | `ABUSEIPDB_API_KEY` |
| `greynoise` | Internet-noise / RIOT classification for an IP | `GREYNOISE_API_KEY` |
| `otx` | AlienVault OTX threat-feed pulse count | `OTX_API_KEY` |
| `securitytrails` | Current DNS + historical A-record changes | `SECURITYTRAILS_API_KEY` |

## How keyless degradation works

If a module's `RequiredKey()` is set and that key is absent from the configuration,
the engine marks the module **`skipped`** (with a human reason like
`"SHODAN_API_KEY not set"`) instead of running or failing it. This means Obscura
Scan produces useful output with **zero keys configured** — the 36 keyless modules
do the heavy lifting.

See [Configuration](CONFIGURATION.md) for the key→provider mapping.
