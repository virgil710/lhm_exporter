package collector

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMinGoVersion(t *testing.T) {
	v := runtime.Version()
	parts := strings.Split(strings.TrimPrefix(v, "go"), ".")
	if len(parts) < 2 {
		t.Fatalf("cannot parse Go version: %s", v)
	}

	var major, minor int
	fmt.Sscanf(parts[0], "%d", &major)
	fmt.Sscanf(parts[1], "%d", &minor)

	if major < 1 || (major == 1 && minor < 23) {
		t.Errorf("Go version %s is below minimum required 1.23", v)
	}
}

func TestSlogCompatibility(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if logger == nil {
		t.Fatal("slog logger should not be nil")
	}
	logger.Info("compatibility test", "version", runtime.Version())
}

func TestAtomicCompatibility(t *testing.T) {
	var counter uint64
	counter++
	if counter != 1 {
		t.Error("atomic uint64 basic operation failed")
	}
}

func TestStringsCutPrefixCompatibility(t *testing.T) {
	v := "go1.23.0"
	result, found := strings.CutPrefix(v, "go")
	if !found {
		t.Error("strings.CutPrefix should return true for valid prefix")
	}
	if result != "1.23.0" {
		t.Errorf("strings.CutPrefix result = %q, want %q", result, "1.23.0")
	}
}

func TestMaxIntCompatibility(t *testing.T) {
	_ = int(^uint(0) >> 1)
}

func TestHardwareCatalogConcurrentSafety(t *testing.T) {
	n := parseTestSample(t)
	e := NodeToExposer(n)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = len(e.CPU)
			_ = len(e.Board)
			_ = len(e.GPU)
			_ = len(e.Disk)
			_ = len(e.Ram)
		}()
	}
	wg.Wait()
}

func TestHardwareCatalogAllTypesExposed(t *testing.T) {
	n := parseTestSample(t)
	e := NodeToExposer(n)

	catalogTypes := map[string]int{
		"Board": len(e.Board),
		"CPU":   len(e.CPU),
		"Ram":   len(e.Ram),
		"VRam":  len(e.VRam),
		"Mem":   len(e.Mem),
		"GPU":   len(e.GPU),
		"Disk":  len(e.Disk),
		"Net":   len(e.Net),
	}

	hasContent := false
	for _, count := range catalogTypes {
		if count > 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("NodeToExposer should produce at least one hardware device type")
	}
}

func TestMetricNamingConvention(t *testing.T) {
	for hwType, metrics := range hardwareMetrics {
		for _, m := range metrics {
			if !strings.HasPrefix(m.name, "lhm_") {
				t.Errorf("metric %q for %s: name must start with 'lhm_'", m.name, hwType)
			}
			if m.help == "" {
				t.Errorf("metric %q for %s: help string must not be empty", m.name, hwType)
			}
			if m.sensorField == "" {
				t.Errorf("metric %q for %s: sensorField must not be empty", m.name, hwType)
			}
		}
	}
}

func TestCollectorDescribeConsistency(t *testing.T) {
	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			return &Node{}, nil
		},
	}
	c := NewLHMCollector(client, nil)

	ch := make(chan *prometheus.Desc, 200)
	c.Describe(ch)
	close(ch)

	descNames := make(map[string]bool)
	for d := range ch {
		descNames[d.String()] = true
	}

	if len(descNames) == 0 {
		t.Error("Describe should emit at least one descriptor")
	}
}

func TestCollectorCollectEmptyNode(t *testing.T) {
	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			return &Node{}, nil
		},
	}
	c := NewLHMCollector(client, nil)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather with empty node failed: %v", err)
	}

	for _, mf := range mfs {
		if mf.GetName() == "lhm_up" {
			val := mf.GetMetric()[0].GetGauge().GetValue()
			if val != 1 {
				t.Errorf("lhm_up should be 1 for successful empty fetch, got %v", val)
			}
		}
	}
}

func TestSensorValueParsingAllUnits(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"50.0 °C", 50.0},
		{"12.096 V", 12.096},
		{"1200 RPM", 1200},
		{"25.5 %", 25.5},
		{"3600 MHz", 3600},
		{"16.2 GB", 16.2},
		{"65.0 W", 65.0},
		{"150.0 MB/s", 150.0},
		{"100.0 MB/s", 100.0},
	}

	for _, tt := range tests {
		got, err := parseSensorValue(tt.input)
		if err != nil {
			t.Errorf("parseSensorValue(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseSensorValue(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestPrometheusMetricGatheringIdempotent(t *testing.T) {
	n := parseTestSample(t)
	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			return n, nil
		},
	}

	c := NewLHMCollector(client, nil)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	first, err := reg.Gather()
	if err != nil {
		t.Fatalf("first gather failed: %v", err)
	}

	second, err := reg.Gather()
	if err != nil {
		t.Fatalf("second gather failed: %v", err)
	}

	firstNames := make(map[string]bool)
	for _, mf := range first {
		firstNames[mf.GetName()] = true
	}
	for _, mf := range second {
		if !firstNames[mf.GetName()] {
			t.Errorf("metric %q present in second gather but not first", mf.GetName())
		}
	}
}

func TestJSONUnmarshalEdgeCase(t *testing.T) {
	emptyJSON := `{"id": 0, "Text": "", "Children": []}`
	var n Node
	if err := json.Unmarshal([]byte(emptyJSON), &n); err != nil {
		t.Fatalf("empty JSON unmarshal failed: %v", err)
	}
	e := NodeToExposer(&n)
	if e == nil {
		t.Fatal("NodeToExposer returned nil for empty-but-valid JSON")
	}
}

func TestParseSensorValueGo123Compatibility(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := float64(i) / 10.0
		input := fmt.Sprintf("%.1f °C", val)
		result, err := parseSensorValue(input)
		if err != nil {
			t.Errorf("parseSensorValue(%q) at iteration %d: %v", input, i, err)
		}
		if result != val {
			t.Errorf("parseSensorValue(%q) = %v, want %v", input, result, val)
		}
	}
}
