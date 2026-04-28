package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"time"

	"lhm_exporter/internal/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/spf13/pflag"
)

var (
	buildTime string
	gitCommit string
	version   string

	metricsPath            = pflag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	listenAddress          = pflag.StringP("web.listen-address", "l", "0.0.0.0", "IP address or host to listen on for web interface and telemetry.")
	listenPort             = pflag.UintP("web.listen-port", "p", 18085, "Port to listen on for web interface and telemetry.")
	disableExporterMetrics = pflag.Bool("web.disable-exporter-metrics", false, "Exclude metrics about the exporter itself (Go runtime and process metrics).")
	showVersion            = pflag.BoolP("version", "v", false, "Show version information.")
	destIP                 = pflag.String("dest.address", "127.0.0.1", "IP address of the monitored device.")
	destPort               = pflag.Uint("dest.port", 8085, "Port of the monitored device.")
	scrapeTimeout          = pflag.Duration("scrape.timeout", 10*time.Second, "Timeout for scraping LHM data.")

	// legacyListen          = pflag.StringP("listen", "l", "", "Deprecated: use --web.listen-address.")
	// legacyListenPort      = pflag.UintP("listen-port", "p", 0, "Deprecated: use --web.listen-address.")
	// legacyDisableGoMetric = pflag.Bool("disable-go-metric", false, "Deprecated: use --web.disable-exporter-metrics.")
)

func main() {
	// mustMarkDeprecated("listen", "use --web.listen-address instead")
	// mustMarkDeprecated("listen-port", "use --web.listen-address instead")
	// mustMarkDeprecated("disable-go-metric", "use --web.disable-exporter-metrics instead")

	pflag.Parse()

	if *showVersion {
		v := version
		if len(v) == 0 {
			v = "dev"
		}
		fmt.Println("lhm_exporter version:", v, "buildTime:", buildTime, "gitCommit:", gitCommit)
		return
	}

	logger := promslog.New(&promslog.Config{})

	// Validate destination IP
	if *destIP != "127.0.0.1" && *destIP != "localhost" {
		if net.ParseIP(*destIP) == nil {
			logger.Error("Target IP address is invalid", "ip", *destIP)
			os.Exit(1)
		}
	}

	// Validate destination port
	if *destPort == 0 || *destPort > 65535 {
		logger.Error("Destination port is invalid", "port", *destPort)
		os.Exit(1)
	}

	if *listenPort == 0 || *listenPort > 65535 {
		logger.Error("Listening port is invalid", "port", *listenPort)
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

	// Create a custom registry to avoid polluting the default global metrics
	reg := prometheus.NewRegistry()

	// Register the LHM collector
	lhmCollector := collector.NewLHMCollector(client, logger)
	reg.MustRegister(lhmCollector)

	// Register Go runtime and process collectors
	if !*disableExporterMetrics {
		reg.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
	}

	// Set up HTTP handlers
	mux := http.NewServeMux()
	mux.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		MaxRequestsInFlight: 2,
	}))

	// Landing page
	if *metricsPath != "/" {
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
			_ = landingTmpl.Execute(w, map[string]string{
				"Version":     v,
				"GitCommit":   gitCommit,
				"BuildTime":   buildTime,
				"MetricsPath": *metricsPath,
			})
		})
	}

	logger.Info("Listening on", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		logger.Error("HTTP server failed", "err", err)
		os.Exit(1)
	}
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
	if len(raw) > 255 {
		return "", fmt.Errorf("host is too long")
	}
	return raw, nil
}

func mustMarkDeprecated(name, message string) {
	if err := pflag.CommandLine.MarkDeprecated(name, message); err != nil {
		panic(err)
	}
}
