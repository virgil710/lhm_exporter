// Package main implements the entry point for the LHM Exporter,
// a Prometheus exporter for LibreHardwareMonitor hardware metrics.
package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"lhm_exporter/internal/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/spf13/pflag"
)

const (
	defaultListenAddr    = "0.0.0.0"
	defaultDestAddr      = "127.0.0.1"
	defaultListenPort    = 18085
	defaultDestPort      = 8085
	defaultScrapeTimeout = 10 * time.Second
	maxRequestsInFlight  = 2
	maxHostnameLength    = 255

	httpReadTimeout       = 10 * time.Second
	httpWriteTimeout      = 30 * time.Second
	httpReadHeaderTimeout = 5 * time.Second
	httpIdleTimeout       = 60 * time.Second
)

var (
	buildTime string
	gitCommit string
	version   string

	metricsPath            = pflag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	listenAddress          = pflag.StringP("web.listen-address", "l", defaultListenAddr, "IP address or host to listen on for web interface and telemetry.")
	listenPort             = pflag.UintP("web.listen-port", "p", defaultListenPort, "Port to listen on for web interface and telemetry.")
	disableExporterMetrics = pflag.Bool("web.disable-exporter-metrics", false, "Exclude metrics about the exporter itself (Go runtime and process metrics).")
	showVersion            = pflag.BoolP("version", "v", false, "Show version information.")
	showHelp               = pflag.BoolP("help", "h", false, "Show help information.")
	destIP                 = pflag.String("dest.address", defaultDestAddr, "IP address of the monitored device.")
	destPort               = pflag.Uint("dest.port", defaultDestPort, "Port of the monitored device.")
	scrapeTimeout          = pflag.Duration("scrape.timeout", defaultScrapeTimeout, "Timeout for scraping LHM data.")
	debugMode              = pflag.Bool("debug", false, "Enable debug mode with verbose logging to stdout.")
)

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage of lhm_exporter:\n")
		pflag.PrintDefaults()
	}

	pflag.Parse()

	if *showVersion || *showHelp {
		if *showVersion {
			v := version
			if len(v) == 0 {
				v = "dev"
			}
			fmt.Println("lhm_exporter version:", v, "buildTime:", buildTime, "gitCommit:", gitCommit)
		}
		if *showHelp {
			fmt.Fprintf(os.Stdout, "Usage of lhm_exporter:\n")
			pflag.PrintDefaults()
		}
		return
	}

	logger := setupLogger(*debugMode)

	if err := validateConfig(); err != nil {
		logger.Error("configuration error", "err", err)
		os.Exit(1)
	}

	listenHost, err := normalizeListenHost(*listenAddress)
	if err != nil {
		logger.Error("Listening address is invalid", "address", *listenAddress, "err", err)
		os.Exit(1)
	}
	listenAddr := net.JoinHostPort(listenHost, fmt.Sprintf("%d", *listenPort))

	v := version
	if len(v) == 0 {
		v = "dev"
	}
	logger.Info("Starting lhm_exporter", "version", v, "destIP", *destIP, "destPort", *destPort, "scrapeTimeout", *scrapeTimeout)

	client := collector.NewLHMClient(*destIP, *destPort, *scrapeTimeout)
	defer client.Close()

	reg := prometheus.NewRegistry()
	lhmCollector := collector.NewLHMCollector(client, logger)
	reg.MustRegister(lhmCollector)

	if !*disableExporterMetrics {
		reg.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
	}

	handler := buildHandler(reg, v, logger)

	logger.Info("Listening on", "address", listenAddr)
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		IdleTimeout:       httpIdleTimeout,
	}
	if err := server.ListenAndServe(); err != nil {
		logger.Error("HTTP server failed", "err", err)
		os.Exit(1)
	}
}

func setupLogger(debugMode bool) *slog.Logger {
	logLevel := promslog.NewLevel()
	if debugMode {
		if err := logLevel.Set("debug"); err != nil {
			fmt.Fprintf(os.Stderr, "failed to set log level: %v\n", err)
		}
	} else {
		if err := logLevel.Set("info"); err != nil {
			fmt.Fprintf(os.Stderr, "failed to set log level: %v\n", err)
		}
	}
	return promslog.New(&promslog.Config{
		Level:  logLevel,
		Format: promslog.NewFormat(),
	})
}

func validateConfig() error {
	if err := validateDestAddress(*destIP); err != nil {
		return fmt.Errorf("dest.address: %w", err)
	}
	if err := validatePort(*destPort, "dest.port"); err != nil {
		return err
	}
	if err := validatePort(*listenPort, "web.listen-port"); err != nil {
		return err
	}
	return nil
}

func validateDestAddress(addr string) error {
	if addr == "localhost" || addr == "127.0.0.1" || addr == "::1" {
		return nil
	}
	if net.ParseIP(addr) != nil {
		return nil
	}
	return fmt.Errorf("invalid IP address: %s", addr)
}

func validatePort(port uint, name string) error {
	if port == 0 || port > 65535 {
		return fmt.Errorf("%s is invalid: %d (must be 1-65535)", name, port)
	}
	return nil
}

func buildHandler(reg *prometheus.Registry, version string, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	cleanPath := path.Clean(*metricsPath)
	if cleanPath == "" {
		cleanPath = "/metrics"
	}

	mux.Handle(cleanPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		MaxRequestsInFlight: maxRequestsInFlight,
	}))

	if cleanPath != "/" {
		landingTmpl := template.Must(template.New("landing").Parse(`<!DOCTYPE html>
<html>
<head><title>LHM Exporter</title></head>
<body>
<h1>LHM Exporter</h1>
<p>Hardware monitoring exporter for LibreHardwareMonitor</p>
<p>Version: {{.Version}}</p>
<p>Commit: {{.GitCommit}}</p>
<p>Built: {{.BuildTime}}</p>
<ul>
  <li><a href="{{.MetricsPath}}">Metrics</a></li>
</ul>
</body>
</html>`))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			if err := landingTmpl.Execute(w, map[string]string{
				"Version":     version,
				"GitCommit":   gitCommit,
				"BuildTime":   buildTime,
				"MetricsPath": *metricsPath,
			}); err != nil {
				logger.Error("failed to render landing page", "err", err)
			}
		})
	}

	var handler http.Handler = mux
	if *debugMode {
		handler = requestLoggingMiddleware(logger, mux)
	}
	return handler
}

// requestLoggingMiddleware wraps an http.Handler and logs each incoming
// request's remote IP and path when debug mode is enabled.
func requestLoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		}
		logger.Debug("HTTP request", "remote_ip", clientIP, "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func normalizeListenHost(raw string) (string, error) {
	if raw == "" {
		return "0.0.0.0", nil
	}
	if _, _, err := net.SplitHostPort(raw); err == nil {
		return "", fmt.Errorf("must not include port; use --web.listen-port separately")
	}
	if raw == "localhost" {
		return raw, nil
	}
	if ip := net.ParseIP(raw); ip != nil {
		return raw, nil
	}
	if len(raw) > maxHostnameLength {
		return "", fmt.Errorf("host is too long")
	}
	return raw, nil
}
