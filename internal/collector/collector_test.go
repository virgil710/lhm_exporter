package collector

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func parseTestSample(t *testing.T) *Node {
	t.Helper()
	var n Node
	if err := json.Unmarshal([]byte(testSampleJSON), &n); err != nil {
		t.Fatalf("Failed to unmarshal test sample: %v", err)
	}
	return &n
}

func TestNodeUnmarshal(t *testing.T) {
	n := parseTestSample(t)
	if n.Text != "Sensor" {
		t.Errorf("Expected root Text='Sensor', got %q", n.Text)
	}
	if len(n.Children) != 1 {
		t.Fatalf("Expected 1 computer child, got %d", len(n.Children))
	}
}

func TestNodeToExposer(t *testing.T) {
	n := parseTestSample(t)
	e := NodeToExposer(n)

	if len(e.Board) == 0 {
		t.Error("Expected at least 1 board")
	}
	if len(e.CPU) == 0 {
		t.Error("Expected at least 1 CPU")
	}
	if len(e.GPU) == 0 {
		t.Error("Expected at least 1 GPU")
	}
	if len(e.Disk) == 0 {
		t.Error("Expected at least 1 disk")
	}
	if len(e.Ram) == 0 {
		t.Error("Expected at least 1 RAM")
	}

	// Verify board device model
	if e.Board[0].DeviceModel != "MSI PRO Z790-P WIFI DDR4 (MS-7E06)" {
		t.Errorf("Unexpected board model: %s", e.Board[0].DeviceModel)
	}
	// Verify board has voltage and temp sensors
	if len(e.Board[0].Voltage) == 0 {
		t.Error("Expected board to have voltage sensors")
	}
	if len(e.Board[0].Temp) == 0 {
		t.Error("Expected board to have temperature sensors")
	}
	if len(e.Board[0].Fan) == 0 {
		t.Error("Expected board to have fan sensors")
	}

	// Verify CPU
	if e.CPU[0].DeviceModel != "Intel Core i7-13700K" {
		t.Errorf("Unexpected CPU model: %s", e.CPU[0].DeviceModel)
	}
	if len(e.CPU[0].Temp) == 0 {
		t.Error("Expected CPU to have temperature sensors")
	}
	if len(e.CPU[0].Load) == 0 {
		t.Error("Expected CPU to have load sensors")
	}
	if len(e.CPU[0].Clock) == 0 {
		t.Error("Expected CPU to have clock sensors")
	}
	if len(e.CPU[0].Power) == 0 {
		t.Error("Expected CPU to have power sensors")
	}

	// Verify GPU
	if e.GPU[0].DeviceModel != "NVIDIA GeForce RTX 4070" {
		t.Errorf("Unexpected GPU model: %s", e.GPU[0].DeviceModel)
	}
	if len(e.GPU[0].Temp) == 0 {
		t.Error("Expected GPU to have temperature sensors")
	}
	if len(e.GPU[0].Fan) == 0 {
		t.Error("Expected GPU to have fan sensors")
	}

	// Verify Disk
	if e.Disk[0].DeviceModel != "Samsung SSD 990 PRO 2TB" {
		t.Errorf("Unexpected disk model: %s", e.Disk[0].DeviceModel)
	}
}

func TestENodeToHardwareDevice(t *testing.T) {
	enode := &ENode{
		baseNode: baseNode{Text: "Test Device"},
		Children: []*ENode{
			{
				baseNode: baseNode{Text: "Temperatures"},
				Children: []*ENode{
					{baseNode: baseNode{Text: "Package"}, Value: "50.0 °C"},
				},
			},
			{
				baseNode: baseNode{Text: "Clocks"},
				Children: []*ENode{
					{baseNode: baseNode{Text: "CPU Core #1"}, Value: "3600 MHz"},
				},
			},
		},
	}

	hd := enode.toHardwareDevice()
	if hd.DeviceModel != "Test Device" {
		t.Errorf("Expected DeviceModel='Test Device', got %q", hd.DeviceModel)
	}
	if len(hd.Temp) != 1 {
		t.Errorf("Expected 1 Temp sensor, got %d", len(hd.Temp))
	}
	if len(hd.Clock) != 1 {
		t.Errorf("Expected 1 Clock sensor, got %d", len(hd.Clock))
	}
}

func TestEmptyNode(t *testing.T) {
	e := NodeToExposer(&Node{})
	if e == nil {
		t.Fatal("NodeToExposer returned nil for empty node")
	}
	if len(e.CPU) != 0 || len(e.Board) != 0 {
		t.Error("Expected empty exposer for empty node")
	}
}

func TestParseSensorValue(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      float64
		wantErr   bool
	}{
		{"temperature", "50.0 °C", 50.0, false},
		{"voltage", "12.096 V", 12.096, false},
		{"fan speed", "1200 RPM", 1200, false},
		{"load percent", "25.5 %", 25.5, false},
		{"clock", "3600 MHz", 3600, false},
		{"empty string", "", 0, true},
		{"non-numeric", "N/A", 0, true},
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

func TestLHMCollectorDescribe(t *testing.T) {
	client := NewLHMClient("127.0.0.1", 8085, 0)
	c := NewLHMCollector(client, nil)

	ch := make(chan *prometheus.Desc, 100)
	c.Describe(ch)

	close(ch)
	count := 0
	for range ch {
		count++
	}
	// Should have at least: lhm_up + scrape metrics + hardware descriptors
	if count < 4 {
		t.Errorf("Expected at least 4 descriptors, got %d", count)
	}
}

func TestLHMCollectorCollectUp(t *testing.T) {
	n := parseTestSample(t)
	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			return n, nil
		},
	}

	c := NewLHMCollector(client, nil)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	// Collect metrics
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify lhm_up is present and = 1
	found := false
	for _, mf := range mfs {
		if mf.GetName() == "lhm_up" {
			found = true
			val := mf.GetMetric()[0].GetGauge().GetValue()
			if val != 1 {
				t.Errorf("Expected lhm_up=1, got %v", val)
			}
		}
	}
	if !found {
		t.Error("lhm_up metric not found")
	}

	// Verify at least some hardware metrics are present
	metricNames := map[string]bool{}
	for _, mf := range mfs {
		metricNames[mf.GetName()] = true
	}

	expectedMetrics := []string{
		"lhm_cpu_temperature_celsius",
		"lhm_cpu_load_percent",
		"lhm_motherboard_voltage_volts",
		"lhm_gpu_temperature_celsius",
		"lhm_disk_temperature_celsius",
	}
	for _, name := range expectedMetrics {
		if !metricNames[name] {
			t.Errorf("Expected metric %s not found", name)
		}
	}

	for _, name := range []string{
		"lhm_scrape_duration_seconds",
		"lhm_scrape_errors_total",
		"lhm_scrape_samples_total",
	} {
		if !metricNames[name] {
			t.Errorf("Expected metric %s not found", name)
		}
	}
}

func TestLHMCollectorCollectDown(t *testing.T) {
	// Create client pointing to non-existent server
	client := NewLHMClient("127.0.0.1", 1, 0)

	c := NewLHMCollector(client, nil)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify lhm_up = 0
	for _, mf := range mfs {
		if mf.GetName() == "lhm_up" {
			val := mf.GetMetric()[0].GetGauge().GetValue()
			if val != 0 {
				t.Errorf("Expected lhm_up=0 when target is unreachable, got %v", val)
			}
		}
	}
}

func TestNewLHMClientHTTPConfig(t *testing.T) {
	client := NewLHMClient("192.168.1.100", 8085, 5)

	if client.url != "http://192.168.1.100:8085/data.json" {
		t.Fatalf("Unexpected client url: %s", client.url)
	}
}

func TestUniqueSensorLabel(t *testing.T) {
	seen := map[string]int{}

	if got := uniqueSensorLabel(seen, "D3D Copy"); got != "D3D Copy" {
		t.Fatalf("first label = %q, want %q", got, "D3D Copy")
	}
	if got := uniqueSensorLabel(seen, "D3D Copy"); got != "D3D Copy #2" {
		t.Fatalf("second label = %q, want %q", got, "D3D Copy #2")
	}
	if got := uniqueSensorLabel(seen, " D3D Copy "); got != "D3D Copy #3" {
		t.Fatalf("trimmed duplicate label = %q, want %q", got, "D3D Copy #3")
	}
	if got := uniqueSensorLabel(seen, " "); got != "unknown" {
		t.Fatalf("blank label = %q, want %q", got, "unknown")
	}
}

func TestLHMCollectorCollectDeduplicatesSensorLabels(t *testing.T) {
	dupJSON := strings.Replace(testSampleJSON, `"Text": "GPU Core", "Min": "0.0 %", "Value": "15.0 %", "Max": "100.0 %", "SensorId": "/gpu/nvidia/0/load/0", "Type": "Load", "ImageURL": "", "Children": null}`,
		`"Text": "D3D Copy", "Min": "0.0 %", "Value": "15.0 %", "Max": "100.0 %", "SensorId": "/gpu/nvidia/0/load/0", "Type": "Load", "ImageURL": "", "Children": null},
                {"id": 290, "Text": "D3D Copy", "Min": "0.0 %", "Value": "10.0 %", "Max": "100.0 %", "SensorId": "/gpu/nvidia/0/load/1", "Type": "Load", "ImageURL": "", "Children": null}`,
		1,
	)
	var n Node
	if err := json.Unmarshal([]byte(dupJSON), &n); err != nil {
		t.Fatalf("failed to parse duplicate test sample: %v", err)
	}

	client := &LHMClient{
		fetchFn: func() (*Node, error) {
			return &n, nil
		},
	}

	c := NewLHMCollector(client, nil)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather with duplicate sensor names failed: %v", err)
	}

	var labels []string
	for _, mf := range mfs {
		if mf.GetName() != "lhm_gpu_load_percent" {
			continue
		}
		for _, metric := range mf.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "sensor_pos" {
					labels = append(labels, label.GetValue())
				}
			}
		}
	}

	foundFirst := false
	foundSecond := false
	for _, label := range labels {
		if label == "D3D Copy" {
			foundFirst = true
		}
		if label == "D3D Copy #2" {
			foundSecond = true
		}
	}

	if !foundFirst || !foundSecond {
		t.Fatalf("expected deduplicated sensor labels, got %v", labels)
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		s    string
		subs []string
		want bool
	}{
		{"/motherboard", []string{"motherboard"}, true},
		{"/cpu/0", []string{"cpu"}, true},
		{"/gpu/nvidia/0", []string{"gpu"}, true},
		{"/nvme/0", []string{"nvme", "hdd", "ssd"}, true},
		{"/ram/0", []string{"ram"}, true},
		{"/vram/0", []string{"vram"}, true},
		{"/unknown/0", []string{"cpu", "gpu"}, false},
		{"", []string{"cpu"}, false},
	}

	for _, tt := range tests {
		got := containsAny(tt.s, tt.subs...)
		if got != tt.want {
			t.Errorf("containsAny(%q, %v) = %v, want %v", tt.s, tt.subs, got, tt.want)
		}
	}
}
