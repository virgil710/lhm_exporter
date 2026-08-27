// Package collector implements a Prometheus collector for LibreHardwareMonitor
// hardware metrics. It fetches data from the LHM web API and exposes them
// as Prometheus gauge metrics.
package collector

import "strings"

const maxRecursionDepth = 20

// baseNode represents the common fields in the LHM JSON hierarchy.
type baseNode struct {
	ID   int    `json:"id"`
	Text string `json:"Text"`
}

// Node is the root of the LHM JSON data structure.
type Node struct {
	baseNode
	Children []ENode `json:"Children"`
}

// ENode represents a hardware device or sensor category in the LHM JSON tree.
// The LHM API uses a recursive JSON structure where the same type appears
// at multiple levels (device, sensor category, individual sensor).
type ENode struct {
	baseNode
	Min        string   `json:"Min"`
	Value      string   `json:"Value"`
	Max        string   `json:"Max"`
	HardwareID string   `json:"HardwareId"`
	SensorID   string   `json:"SensorId"`
	Type       string   `json:"Type"`
	Children   []*ENode `json:"Children"`
}

// HardwareDevice groups sensors by type for a single hardware device.
type HardwareDevice struct {
	DeviceModel string
	Voltage     []*ENode
	Temp        []*ENode
	Fan         []*ENode
	Control     []*ENode
	Power       []*ENode
	Clock       []*ENode
	Load        []*ENode
	Data        []*ENode
	Timing      []*ENode
	Throughput  []*ENode
	Level       []*ENode
	Factor      []*ENode
}

// HardwareCatalog categorizes hardware devices by type.
type HardwareCatalog struct {
	Board []HardwareDevice
	CPU   []HardwareDevice
	Ram   []HardwareDevice
	VRam  []HardwareDevice
	Mem   []HardwareDevice
	GPU   []HardwareDevice
	Disk  []HardwareDevice
	Net   []HardwareDevice
}

// toHardwareDevice converts an ENode (hardware device) into a HardwareDevice
// by recursively collecting sensor categories from all descendant levels.
// The LHM tree structure is: device → chip → sensor category → sensor.
func (e *ENode) toHardwareDevice() *HardwareDevice {
	hd := &HardwareDevice{
		DeviceModel: e.Text,
	}
	collectSensors(e, hd, 0)
	return hd
}

// collectSensors recursively walks the ENode tree and collects
// sensor leaf nodes into the appropriate HardwareDevice fields.
func collectSensors(node *ENode, hd *HardwareDevice, depth int) {
	if depth > maxRecursionDepth {
		return
	}
	for _, child := range node.Children {
		switch child.Text {
		case "Voltages":
			hd.Voltage = append(hd.Voltage, child.Children...)
		case "Temperatures":
			hd.Temp = append(hd.Temp, child.Children...)
		case "Fans":
			hd.Fan = append(hd.Fan, child.Children...)
		case "Controls":
			hd.Control = append(hd.Control, child.Children...)
		case "Powers":
			hd.Power = append(hd.Power, child.Children...)
		case "Clocks":
			hd.Clock = append(hd.Clock, child.Children...)
		case "Load":
			hd.Load = append(hd.Load, child.Children...)
		case "Data":
			hd.Data = append(hd.Data, child.Children...)
		case "Timings":
			hd.Timing = append(hd.Timing, child.Children...)
		case "Throughput":
			hd.Throughput = append(hd.Throughput, child.Children...)
		case "Level":
			hd.Level = append(hd.Level, child.Children...)
		case "Factors":
			hd.Factor = append(hd.Factor, child.Children...)
		default:
			collectSensors(child, hd, depth+1)
		}
	}
}

// NodeToExposer converts the raw LHM JSON tree into a categorized HardwareCatalog.
func NodeToExposer(n *Node) *HardwareCatalog {
	e := &HardwareCatalog{}
	for _, computer := range n.Children {
		for _, device := range computer.Children {
			hd := device.toHardwareDevice()
			switch {
			case containsAny(device.HardwareID, "motherboard"):
				e.Board = append(e.Board, *hd)
			case containsAny(device.HardwareID, "cpu"):
				e.CPU = append(e.CPU, *hd)
			case containsAny(device.HardwareID, "vram"):
				e.VRam = append(e.VRam, *hd)
			case containsAny(device.HardwareID, "ram"):
				e.Ram = append(e.Ram, *hd)
			case containsAny(device.HardwareID, "memory"):
				e.Mem = append(e.Mem, *hd)
			case containsAny(device.HardwareID, "gpu"):
				e.GPU = append(e.GPU, *hd)
			case containsAny(device.HardwareID, "nic"):
				e.Net = append(e.Net, *hd)
			case containsAny(device.HardwareID, "nvme", "hdd", "ssd"):
				e.Disk = append(e.Disk, *hd)
			}
		}
	}
	return e
}

// containsAny reports whether s contains any of the given substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
