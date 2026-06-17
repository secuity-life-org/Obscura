// Command obscura is the Obscura Scan threat-hunter & attack-surface platform.
//
// It loads config, opens the WAL database (importing a legacy aegis.db if
// present), builds the shared SSRF-guarded HTTP client and scan engine, and
// serves the web UI + API with graceful shutdown.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"obscurascan/internal/config"
	"obscurascan/internal/engine"
	"obscurascan/internal/httpx"
	"obscurascan/internal/schedule"
	"obscurascan/internal/server"
	"obscurascan/internal/store"

	// Register all scan modules via their init() functions.
	_ "obscurascan/internal/modules"
)

const banner = `
  ___  _                            ___
 / _ \| |__ ___ __ _   _ _ _ __ _  / __| __ __ _ _ _
| (_) | '_ (_-</ _| || | '_/ _` + "`" + ` | \__ \/ _/ _` + "`" + ` | ' \
 \___/|_.__/__/\__|\_,_|_| \__,_| |___/\__\__,_|_||_|
       Obscura Scan — Threat Hunter & Attack Surface Management
`

func main() {
	var showVersion, allowInternal bool
	var mintKey, mintRole string
	flag.BoolVar(&showVersion, "version", false, "print build info and exit")
	flag.BoolVar(&allowInternal, "allow-internal", false, "permit scanning private/loopback targets (default deny)")
	flag.StringVar(&mintKey, "mint-key", "", "create an API key with the given name, print it, and exit")
	flag.StringVar(&mintRole, "mint-role", "viewer", "role for --mint-key (viewer|admin)")
	flag.Parse()

	if showVersion {
		fmt.Printf("Obscura Scan %s (commit %s, built %s)\n", version, commit, buildDate)
		return
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	fmt.Fprint(os.Stderr, banner)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "err", err)
		os.Exit(1)
	}
	if allowInternal {
		cfg.AllowInternal = true
	}

	slog.Info("configuration loaded",
		"version", cfg.Version, "host", cfg.Host, "port", cfg.Port,
		"db_path", cfg.DBPath, "modules", len(engine.Names()),
		"configured_api_keys", cfg.ConfiguredAPIKeys(),
		"allow_internal", cfg.AllowInternal,
	)

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer st.Close()
	slog.Info("database ready (WAL)", "path", cfg.DBPath)
	seedTemplates(st)

	// --mint-key: create an API key, print the plaintext once, and exit.
	if mintKey != "" {
		b := make([]byte, 24)
		_, _ = rand.Read(b)
		plain := "obx_" + hex.EncodeToString(b)
		sum := sha256.Sum256([]byte(plain))
		if _, err := st.APIKeys().Create(mintKey, hex.EncodeToString(sum[:]), mintRole); err != nil {
			slog.Error("failed to mint API key", "err", err)
			os.Exit(1)
		}
		fmt.Printf("API key '%s' (%s) created. Store it now — it is not recoverable:\n\n  %s\n\n", mintKey, mintRole, plain)
		return
	}

	// Root context cancelled on SIGINT/SIGTERM; in-flight scans tie to it.
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := httpx.New(httpx.Options{
		AllowInternal: cfg.AllowInternal,
		Timeout:       time.Duration(cfg.DefaultTimeout) * time.Second,
	})

	runner := engine.NewRunner(st, cfg, client)

	// Background scheduler for recurring scans.
	go schedule.New(st, runner).Start(rootCtx)

	srv, err := server.New(rootCtx, cfg, st, client, runner)
	if err != nil {
		slog.Error("failed to build server", "err", err)
		os.Exit(1)
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("Obscura Scan listening", "url", "http://"+addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "err", err)
			stop()
		}
	}()

	<-rootCtx.Done()
	slog.Info("shutdown signal received — draining (max 20s)")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	}
	slog.Info("Obscura Scan stopped")
}

// seedTemplates installs the default scan profiles once (idempotent upsert).
func seedTemplates(st *store.Store) {
	defaults := []struct {
		name, desc, mode string
		modules          []string
	}{
		{"Quick", "Fast baseline: DNS, TLS, headers, exposed files.", "defensive",
			[]string{"dns_records", "tls", "http_probe", "sec_headers", "tech"}},
		{"Full", "Broad keyless recon across all categories.", "defensive",
			[]string{"dns_records", "whois", "tls", "ssl_chain", "sec_headers", "tech", "http_probe", "waf_detect",
				"cert_transparency", "subdomain_scan", "spf_analyzer", "ip_geolocation", "crawler", "robots_txt",
				"security_txt", "http_methods", "cors", "cookie_audit", "wayback_urls", "reverse_ip"}},
		{"Bug Bounty", "Attack-surface discovery for bug hunters.", "semi-offensive",
			[]string{"subdomain_scan", "subdomain_permutation", "http_probe", "cors", "http_methods",
				"wayback_urls", "google_dorking", "cert_transparency", "dns_zone_transfer", "favicon_pivot"}},
		{"Compliance", "Posture checks for audits (headers, TLS, email).", "defensive",
			[]string{"sec_headers", "ssl_chain", "tls", "spf_analyzer", "cookie_audit", "http_methods", "security_txt"}},
	}
	if st.Templates().Count() > 0 {
		return
	}
	for _, d := range defaults {
		_ = st.Templates().Upsert(d.name, d.desc, d.modules, d.mode)
	}
	slog.Info("seeded default scan templates", "count", len(defaults))
}
