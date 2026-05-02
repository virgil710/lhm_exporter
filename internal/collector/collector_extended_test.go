package collector

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestParseSensorValueEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"negative temperature", "-5.0 °C", -5.0, false},
		{"scientific notation", "1e3 MHz", 1000, false},
		{"large value", "1000000000 Hz", 1000000000, false},
		{"decimal", "16.2 GB", 16.2, false},
		{"small voltage", "0.001 V", 0.001, false},
		{"multiple dots", "123.456.789", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSensorValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSensorValue(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSensorValue(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMalformedJSONHandling(t *testing.T) {
	malformedJSON := `{invalid json content`

	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			var n Node
			if err := json.Unmarshal([]byte(malformedJSON), &n); err != nil {
				return nil, err
			}
			return &n, nil
		},
	}

	c := NewLHMCollector(client, nil)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, mf := range mfs {
		if mf.GetName() == "lhm_up" {
			val := mf.GetMetric()[0].GetGauge().GetValue()
			if val != 0 {
				t.Errorf("Expected lhm_up=0 for malformed JSON, got %v", val)
			}
		}
	}
}

func TestVRAMHardwareType(t *testing.T) {
	vramJSON := `{
		"id": 0,
		"Text": "Sensor",
		"Children": [
			{
				"id": 1,
				"Text": "Computer",
				"Children": [
					{
						"id": 2,
						"Text": "VRAM Device",
						"HardwareId": "/vram/0",
						"Children": [
							{
								"id": 3,
								"Text": "Load",
								"Children": [
									{"id": 4, "Text": "VRAM Usage", "Value": "75.5 %", "Type": "Load"}
								]
							}
						]
					}
				]
			}
		]
	}`

	var n Node
	if err := json.Unmarshal([]byte(vramJSON), &n); err != nil {
		t.Fatalf("Failed to unmarshal VRAM test data: %v", err)
	}

	e := NodeToExposer(&n)
	if len(e.VRam) == 0 {
		t.Error("Expected VRAM device to be detected")
	}
}

func TestPhysicalMemoryHardwareType(t *testing.T) {
	memJSON := `{
		"id": 0,
		"Text": "Sensor",
		"Children": [
			{
				"id": 1,
				"Text": "Computer",
				"Children": [
					{
						"id": 2,
						"Text": "Memory Device",
						"HardwareId": "/memory/0",
						"Children": [
							{
								"id": 3,
								"Text": "Data",
								"Children": [
									{"id": 4, "Text": "Total", "Value": "32.0 GB", "Type": "Data"}
								]
							}
						]
					}
				]
			}
		]
	}`

	var n Node
	if err := json.Unmarshal([]byte(memJSON), &n); err != nil {
		t.Fatalf("Failed to unmarshal memory test data: %v", err)
	}

	e := NodeToExposer(&n)
	if len(e.Mem) == 0 {
		t.Error("Expected physical memory device to be detected")
	}
}

func TestNetworkInterfaceHardwareType(t *testing.T) {
	netJSON := `{
		"id": 0,
		"Text": "Sensor",
		"Children": [
			{
				"id": 1,
				"Text": "Computer",
				"Children": [
					{
						"id": 2,
						"Text": "Network Interface",
						"HardwareId": "/nic/0",
						"Children": [
							{
								"id": 3,
								"Text": "Throughput",
								"Children": [
									{"id": 4, "Text": "Download", "Value": "100.0 MB/s", "Type": "Throughput"}
								]
							}
						]
					}
				]
			}
		]
	}`

	var n Node
	if err := json.Unmarshal([]byte(netJSON), &n); err != nil {
		t.Fatalf("Failed to unmarshal network test data: %v", err)
	}

	e := NodeToExposer(&n)
	if len(e.Net) == 0 {
		t.Error("Expected network interface to be detected")
	}
}

func TestCollectorWithNilLogger(t *testing.T) {
	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			return &Node{}, nil
		},
	}

	c := NewLHMCollector(client, nil)
	if c.logger == nil {
		t.Error("Logger should not be nil when nil is passed")
	}
}

func TestScrapeDurationMetric(t *testing.T) {
	n := parseTestSample(t)
	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			return n, nil
		},
	}

	c := NewLHMCollector(client, nil)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "lhm_scrape_duration_seconds" {
			found = true
			val := mf.GetMetric()[0].GetGauge().GetValue()
			if val < 0 {
				t.Errorf("Expected non-negative scrape duration, got %v", val)
			}
		}
	}
	if !found {
		t.Error("lhm_scrape_duration_seconds metric not found")
	}
}

func TestScrapeErrorsCounterIncrements(t *testing.T) {
	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	c := NewLHMCollector(client, nil)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	_, _ = reg.Gather()

	mfs, _ := reg.Gather()
	for _, mf := range mfs {
		if mf.GetName() == "lhm_scrape_errors_total" {
			val := mf.GetMetric()[0].GetCounter().GetValue()
			if val == 0 {
				t.Error("Expected scrape errors counter to be non-zero after failed fetch")
			}
		}
	}
}

func TestHardwareIdCaseSensitivity(t *testing.T) {
	tests := []struct {
		name       string
		hardwareId string
		expected   string
	}{
		{"lowercase cpu", "/cpu/0", "cpu"},
		{"uppercase CPU", "/CPU/0", ""},
		{"mixed case", "/Cpu/0", ""},
		{"nvme lowercase", "/nvme/0", "disk"},
		{"NVME uppercase", "/NVME/0", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := `{
				"id": 0,
				"Text": "Sensor",
				"Children": [{
					"id": 1,
					"Text": "Computer",
					"Children": [{
						"id": 2,
						"Text": "Device",
						"HardwareId": "` + tt.hardwareId + `",
						"Children": [{
							"id": 3,
							"Text": "Temperatures",
							"Children": [{"id": 4, "Text": "Temp", "Value": "50.0 °C", "Type": "Temperature"}]
						}]
					}]
				}]
			}`

			var n Node
			if err := json.Unmarshal([]byte(jsonStr), &n); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			e := NodeToExposer(&n)

			switch tt.expected {
			case "cpu":
				if len(e.CPU) == 0 {
					t.Errorf("Expected CPU device for HardwareId %q", tt.hardwareId)
				}
			case "disk":
				if len(e.Disk) == 0 {
					t.Errorf("Expected disk device for HardwareId %q", tt.hardwareId)
				}
			case "":
				if len(e.CPU) > 0 || len(e.Disk) > 0 {
					t.Errorf("Expected no device match for HardwareId %q", tt.hardwareId)
				}
			}
		})
	}
}

func TestContainsAnyEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		s    string
		subs []string
		want bool
	}{
		{"empty string", "", []string{"a"}, false},
		{"empty substrings", "test", []string{}, false},
		{"empty both", "", []string{}, false},
		{"substring at start", "cpu/0", []string{"cpu"}, true},
		{"substring at end", "/cpu", []string{"cpu"}, true},
		{"substring in middle", "/cpu/0", []string{"cpu"}, true},
		{"no match", "/gpu/0", []string{"cpu"}, false},
		{"multiple subs one match", "/cpu/0", []string{"gpu", "cpu", "ram"}, true},
		{"case sensitive no match", "/CPU/0", []string{"cpu"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAny(tt.s, tt.subs...)
			if got != tt.want {
				t.Errorf("containsAny(%q, %v) = %v, want %v", tt.s, tt.subs, got, tt.want)
			}
		})
	}
}
