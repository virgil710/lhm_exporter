package collector

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (int, error) { return len(p), nil }

const (
	namespace = "lhm"
)

var labelNames = []string{"device", "device_model", "sensor_pos"}

// metricDef defines a Prometheus metric descriptor specification.
type metricDef struct {
	sensorField string // field name in HardwareDevice (e.g., "Temp", "Voltage")
	name        string // full metric name (e.g., "lhm_cpu_temperature_celsius")
	help        string
}

// hardwareMetrics maps hardware type prefixes to their metric definitions.
var hardwareMetrics = map[string][]metricDef{
	"cpu": {
		{sensorField: "Temp", name: "lhm_cpu_temperature_celsius", help: "CPU temperature in celsius, from LibreHardwareMonitor"},
		{sensorField: "Voltage", name: "lhm_cpu_voltage_volts", help: "CPU voltage in volts, from LibreHardwareMonitor"},
		{sensorField: "Power", name: "lhm_cpu_power_watts", help: "CPU power in watts, from LibreHardwareMonitor"},
		{sensorField: "Clock", name: "lhm_cpu_clock_hertz", help: "CPU clock frequency in hertz, from LibreHardwareMonitor"},
		{sensorField: "Load", name: "lhm_cpu_load_percent", help: "CPU load in percent, from LibreHardwareMonitor"},
	},
	"motherboard": {
		{sensorField: "Temp", name: "lhm_motherboard_temperature_celsius", help: "Motherboard temperature in celsius, from LibreHardwareMonitor"},
		{sensorField: "Voltage", name: "lhm_motherboard_voltage_volts", help: "Motherboard voltage in volts, from LibreHardwareMonitor"},
		{sensorField: "Fan", name: "lhm_motherboard_fan_speed_rpm", help: "Motherboard fan speed in RPM, from LibreHardwareMonitor"},
		{sensorField: "Control", name: "lhm_motherboard_control_percent", help: "Motherboard control in percent, from LibreHardwareMonitor"},
	},
	"ram": {
		{sensorField: "Load", name: "lhm_ram_load_percent", help: "RAM load in percent, from LibreHardwareMonitor"},
		{sensorField: "Data", name: "lhm_ram_data_bytes", help: "RAM data in bytes, from LibreHardwareMonitor"},
	},
	"vram": {
		{sensorField: "Load", name: "lhm_vram_load_percent", help: "VRAM load in percent, from LibreHardwareMonitor"},
		{sensorField: "Data", name: "lhm_vram_data_bytes", help: "VRAM data in bytes, from LibreHardwareMonitor"},
	},
	"physical_memory": {
		{sensorField: "Data", name: "lhm_physical_memory_data_bytes", help: "Physical memory data in bytes, from LibreHardwareMonitor"},
		{sensorField: "Timing", name: "lhm_physical_memory_timing_nanoseconds", help: "Physical memory timing in nanoseconds, from LibreHardwareMonitor"},
	},
	"gpu": {
		{sensorField: "Temp", name: "lhm_gpu_temperature_celsius", help: "GPU temperature in celsius, from LibreHardwareMonitor"},
		{sensorField: "Voltage", name: "lhm_gpu_voltage_volts", help: "GPU voltage in volts, from LibreHardwareMonitor"},
		{sensorField: "Power", name: "lhm_gpu_power_watts", help: "GPU power in watts, from LibreHardwareMonitor"},
		{sensorField: "Clock", name: "lhm_gpu_clock_hertz", help: "GPU clock frequency in hertz, from LibreHardwareMonitor"},
		{sensorField: "Load", name: "lhm_gpu_load_percent", help: "GPU load in percent, from LibreHardwareMonitor"},
		{sensorField: "Fan", name: "lhm_gpu_fan_speed_rpm", help: "GPU fan speed in RPM, from LibreHardwareMonitor"},
		{sensorField: "Control", name: "lhm_gpu_control_percent", help: "GPU control in percent, from LibreHardwareMonitor"},
		{sensorField: "Data", name: "lhm_gpu_data_bytes", help: "GPU data in bytes, from LibreHardwareMonitor"},
		{sensorField: "Throughput", name: "lhm_gpu_throughput_bytes_per_second", help: "GPU throughput in bytes per second, from LibreHardwareMonitor"},
	},
	"disk": {
		{sensorField: "Temp", name: "lhm_disk_temperature_celsius", help: "Disk temperature in celsius, from LibreHardwareMonitor"},
		{sensorField: "Load", name: "lhm_disk_load_percent", help: "Disk load in percent, from LibreHardwareMonitor"},
		{sensorField: "Level", name: "lhm_disk_level_percent", help: "Disk level in percent, from LibreHardwareMonitor"},
		{sensorField: "Factor", name: "lhm_disk_factor_ratio", help: "Disk factor ratio (dimensionless), from LibreHardwareMonitor"},
		{sensorField: "Data", name: "lhm_disk_data_bytes", help: "Disk data in bytes, from LibreHardwareMonitor"},
		{sensorField: "Throughput", name: "lhm_disk_throughput_bytes_per_second", help: "Disk throughput in bytes per second, from LibreHardwareMonitor"},
	},
	"net": {
		{sensorField: "Data", name: "lhm_net_data_bytes", help: "Network data in bytes, from LibreHardwareMonitor"},
		{sensorField: "Load", name: "lhm_net_load_percent", help: "Network load in percent, from LibreHardwareMonitor"},
		{sensorField: "Throughput", name: "lhm_net_throughput_bytes_per_second", help: "Network throughput in bytes per second, from LibreHardwareMonitor"},
	},
}

// LHMCollector implements the prometheus.Collector interface for
// LibreHardwareMonitor hardware metrics.
type LHMCollector struct {
	client *LHMClient
	logger *slog.Logger

	descs map[string]*prometheus.Desc
	up    *prometheus.Desc

	scrapeDuration *prometheus.Desc
	scrapeErrors   *prometheus.Desc
	scrapeSamples  *prometheus.Desc

	errorCount uint64
}

// NewLHMCollector creates a new LHMCollector.
func NewLHMCollector(client *LHMClient, logger *slog.Logger) *LHMCollector {
	if logger == nil {
		logger = newDiscardLogger()
	}
	descs := make(map[string]*prometheus.Desc)
	for _, metrics := range hardwareMetrics {
		for _, m := range metrics {
			if _, exists := descs[m.name]; !exists {
				descs[m.name] = prometheus.NewDesc(m.name, m.help, labelNames, nil)
			}
		}
	}

	return &LHMCollector{
		client: client,
		logger: logger,
		descs:  descs,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Was the last scrape of LibreHardwareMonitor successful.",
			nil, nil,
		),
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
			"Duration of the last scrape in seconds.",
			nil, nil,
		),
		scrapeErrors: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_errors_total"),
			"Total number of scrape errors.",
			nil, nil,
		),
		scrapeSamples: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_samples_total"),
			"Total number of samples scraped during the last scrape.",
			nil, nil,
		),
	}
}

// Describe sends all metric descriptors to the channel.
func (c *LHMCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
	ch <- c.up
	ch <- c.scrapeDuration
	ch <- c.scrapeErrors
	ch <- c.scrapeSamples
}

// Collect fetches data from LHM and sends all metrics to the channel.
func (c *LHMCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	samples := 0

	node, err := c.client.Fetch()
	if err != nil {
		c.logger.Error("failed to fetch LHM data", "err", err)
		atomic.AddUint64(&c.errorCount, 1)
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
		ch <- prometheus.MustNewConstMetric(c.scrapeErrors, prometheus.CounterValue, float64(atomic.LoadUint64(&c.errorCount)))
		ch <- prometheus.MustNewConstMetric(c.scrapeSamples, prometheus.GaugeValue, 0)
		return
	}

	exposer := NodeToExposer(node)

	// Collect metrics for each hardware type
	samples += c.collectHardware(ch, "cpu", exposer.CPU, "cpu")
	samples += c.collectHardware(ch, "motherboard", exposer.Board, "board")
	samples += c.collectHardware(ch, "ram", exposer.Ram, "ram")
	samples += c.collectHardware(ch, "vram", exposer.VRam, "vram")
	samples += c.collectHardware(ch, "physical_memory", exposer.Mem, "physical_memory")
	samples += c.collectHardware(ch, "gpu", exposer.GPU, "gpu")
	samples += c.collectHardware(ch, "disk", exposer.Disk, "disk")
	samples += c.collectHardware(ch, "net", exposer.Net, "net")

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
	ch <- prometheus.MustNewConstMetric(c.scrapeErrors, prometheus.CounterValue, float64(atomic.LoadUint64(&c.errorCount)))
	ch <- prometheus.MustNewConstMetric(c.scrapeSamples, prometheus.GaugeValue, float64(samples))

	c.logger.Debug("collected LHM metrics", "duration", time.Since(start))
}

// collectHardware emits metrics for a specific hardware type.
func (c *LHMCollector) collectHardware(ch chan<- prometheus.Metric, hwType string, devices []HardwareDevice, devicePrefix string) int {
	metrics, ok := hardwareMetrics[hwType]
	if !ok {
		return 0
	}

	samples := 0
	for idx, device := range devices {
		deviceLabel := devicePrefix + strconv.Itoa(idx)
		for _, m := range metrics {
			sensors := getSensorsByField(&device, m.sensorField)
			desc, ok := c.descs[m.name]
			if !ok {
				continue
			}
			sensorLabels := make(map[string]int, len(sensors))
			for _, s := range sensors {
				value := parseSensorValue(s.Value)
				sensorLabel := uniqueSensorLabel(sensorLabels, s.Text)
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, deviceLabel, device.DeviceModel, sensorLabel)
				samples++
			}
		}
	}
	return samples
}

func uniqueSensorLabel(seen map[string]int, raw string) string {
	label := strings.TrimSpace(raw)
	if label == "" {
		label = "unknown"
	}
	seen[label]++
	if seen[label] == 1 {
		return label
	}
	return fmt.Sprintf("%s #%d", label, seen[label])
}

// getSensorsByField returns the sensor slice corresponding to the given field name.
func getSensorsByField(hd *HardwareDevice, field string) []*ENode {
	switch field {
	case "Temp":
		return hd.Temp
	case "Voltage":
		return hd.Voltage
	case "Fan":
		return hd.Fan
	case "Control":
		return hd.Control
	case "Power":
		return hd.Power
	case "Clock":
		return hd.Clock
	case "Load":
		return hd.Load
	case "Data":
		return hd.Data
	case "Timing":
		return hd.Timing
	case "Throughput":
		return hd.Throughput
	case "Level":
		return hd.Level
	case "Factor":
		return hd.Factor
	default:
		return nil
	}
}

// parseSensorValue extracts the numeric value from a sensor string like "50.0 °C".
// Returns -1 if parsing fails.
func parseSensorValue(s string) float64 {
	if s == "" {
		return -1
	}
	v := strings.SplitN(s, " ", 2)[0]
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return -1
	}
	return f
}
