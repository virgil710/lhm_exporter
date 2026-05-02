# LHM Exporter

[![Go CI](https://github.com/milkbrother666/lhm_exporter_private/actions/workflows/go.yml/badge.svg)](https://github.com/milkbrother666/lhm_exporter_private/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.23-00ADD8.svg)](https://go.dev/)

A Prometheus exporter for [LibreHardwareMonitor](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor), exposing Windows hardware metrics (CPU, GPU, disk, memory, network, etc.).

基于 [LibreHardwareMonitor](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor) 的 Prometheus Exporter，支持采集 Windows 硬件指标（CPU、GPU、磁盘、内存、网络等）。

## Table of Contents

- [Prerequisites](#prerequisites--前置条件)
- [Quick Start](#quick-start--快速上手)
- [Installation](#installation--安装)
- [Usage](#usage--使用方法)
- [Prometheus Configuration](#prometheus-configuration--prometheus-配置示例)
- [Metrics](#metrics--指标列表)
- [Notice](#notice--注意事项)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

## Prerequisites / 前置条件

[LibreHardwareMonitor](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor) needs to be installed on the monitored machine, and its web interface must be enabled.

需要在被监控的机器上安装 LibreHardwareMonitor 并启用 Web 接口：

Option -> Remote Web Server -> Run

![](.readme/lhm_exporter01.jpg)

## Quick Start / 快速上手

1. 在被监控 Windows 机器上安装 [LibreHardwareMonitor](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor)，启用 Web Server（默认端口 `8085`）。
2. 从 [Releases](https://github.com/milkbrother666/lhm_exporter_private/releases) 页面下载对应平台的二进制文件，或从源码构建：
   ```bash
   make build
   ```
3. 启动 exporter（监控本机）：
   ```bash
   ./lhm_exporter
   ```
4. 访问 `http://localhost:18085/metrics` 验证指标是否正常采集。

## Installation / 安装

### Pre-built binaries / 预编译二进制

Download the latest release from the [Releases](https://github.com/milkbrother666/lhm_exporter_private/releases) page.

从 [Releases](https://github.com/milkbrother666/lhm_exporter_private/releases) 页面下载最新版本。支持以下平台：

| Platform       | Arch       |
| -------------- | ---------- |
| Linux          | amd64, arm64 |
| macOS (Darwin) | amd64, arm64 |
| Windows        | amd64      |

### Build from source / 源码构建

Requires Go >= 1.23.

```bash
git clone https://github.com/milkbrother666/lhm_exporter_private.git
cd lhm_exporter_private
make build
```

## Usage / 使用方法

`lhm_exporter` does not need to be deployed on the same machine as LibreHardwareMonitor; it only needs to be able to access its interface over the network.

`lhm_exporter` 不需要与 LibreHardwareMonitor 部署在同一台机器上，只需能通过网络访问其接口即可。

```bash
lhm_exporter [flags]
```

### Flags / 启动参数

| Flag                             | Short | Default     | Description                                                     |
| -------------------------------- | ----- | ----------- | --------------------------------------------------------------- |
| `--web.listen-address`           | `-l`  | `0.0.0.0`   | IP address or host to listen on for web interface and telemetry |
| `--web.listen-port`              | `-p`  | `18085`     | Port to listen on for web interface and telemetry               |
| `--web.telemetry-path`           |       | `/metrics`  | Path under which to expose metrics                              |
| `--web.disable-exporter-metrics` |       | `false`     | Exclude metrics about the exporter itself (Go runtime and process metrics) |
| `--dest.address`                 |       | `127.0.0.1` | IP address of the monitored device                              |
| `--dest.port`                    |       | `8085`      | Port of the monitored device                                    |
| `--scrape.timeout`               |       | `10s`       | Timeout for scraping LHM data                                   |
| `--debug`                        |       | `false`     | Enable debug mode with verbose logging to stdout                |
| `--version`                      | `-v`  |             | Show version information                                        |
| `--help`                         | `-h`  |             | Show help information                                           |

### Examples / 常用用法

Local machine monitoring / 监控本地机器：

```bash
lhm_exporter
```

Remote machine monitoring / 监控远程机器：

```bash
lhm_exporter --dest.address 192.168.1.100
```

Custom listen port / 自定义监听端口：

```bash
lhm_exporter -p 9100
```

LibreHardwareMonitor's built-in web interface is HTTP only, so `lhm_exporter` always connects to the upstream target over HTTP.

LibreHardwareMonitor 内置 Web 接口仅支持 HTTP，因此 `lhm_exporter` 始终通过 HTTP 连接上游目标。

The default port is 18085, and the metrics endpoint is: `http://<your-ip>:18085/metrics`.

默认端口为 18085，指标访问地址为：`http://<your-ip>:18085/metrics`。

## Prometheus Configuration / Prometheus 配置示例

```yaml
scrape_configs:
  - job_name: 'lhm'
    static_configs:
      - targets: ['localhost:18085']
    metrics_path: /metrics
    scrape_interval: 15s
    scrape_timeout: 10s
```

## Metrics / 指标列表

### Exporter health / Exporter 自身指标

| Metric                        | Type    | Description                                            |
| ----------------------------- | ------- | ------------------------------------------------------ |
| `lhm_up`                      | Gauge   | Was the last scrape of LibreHardwareMonitor successful |
| `lhm_scrape_duration_seconds` | Gauge   | Duration of the last scrape in seconds                 |
| `lhm_scrape_errors_total`     | Counter | Total number of scrape errors                          |
| `lhm_scrape_samples_total`    | Gauge   | Total number of samples scraped                        |

### Hardware metrics / 硬件指标

All hardware metrics are of type `Gauge` and share the following labels:

所有硬件指标均为 `Gauge` 类型，共享以下标签：

| Label          | Description                        |
| -------------- | ---------------------------------- |
| `device`       | Hardware device identifier (e.g., `cpu0`, `gpu0`) |
| `device_model` | Hardware model name from LHM       |
| `sensor_pos`   | Sensor position within the device  |

| Metric                                   | Description                       |
| ---------------------------------------- | --------------------------------- |
| `lhm_cpu_temperature_celsius`            | CPU temperature (°C)              |
| `lhm_cpu_voltage_volts`                  | CPU voltage (V)                   |
| `lhm_cpu_power_watts`                    | CPU power consumption (W)         |
| `lhm_cpu_clock_hertz`                    | CPU clock frequency (Hz)          |
| `lhm_cpu_load_percent`                   | CPU load (%)                      |
| `lhm_motherboard_temperature_celsius`    | Motherboard temperature (°C)      |
| `lhm_motherboard_voltage_volts`          | Motherboard voltage (V)           |
| `lhm_motherboard_fan_speed_rpm`          | Motherboard fan speed (RPM)       |
| `lhm_motherboard_control_percent`        | Motherboard control (%)           |
| `lhm_ram_load_percent`                   | RAM load (%)                      |
| `lhm_ram_data_bytes`                     | RAM data (bytes)                  |
| `lhm_vram_load_percent`                  | VRAM load (%)                     |
| `lhm_vram_data_bytes`                    | VRAM data (bytes)                 |
| `lhm_physical_memory_data_bytes`         | Physical memory data (bytes)      |
| `lhm_physical_memory_timing_nanoseconds` | Physical memory timing (ns)       |
| `lhm_gpu_temperature_celsius`            | GPU temperature (°C)              |
| `lhm_gpu_voltage_volts`                  | GPU voltage (V)                   |
| `lhm_gpu_power_watts`                    | GPU power consumption (W)         |
| `lhm_gpu_clock_hertz`                    | GPU clock frequency (Hz)          |
| `lhm_gpu_load_percent`                   | GPU load (%)                      |
| `lhm_gpu_fan_speed_rpm`                  | GPU fan speed (RPM)               |
| `lhm_gpu_control_percent`                | GPU control (%)                   |
| `lhm_gpu_data_bytes`                     | GPU data (bytes)                  |
| `lhm_gpu_throughput_bytes_per_second`    | GPU throughput (bytes/s)          |
| `lhm_disk_temperature_celsius`           | Disk temperature (°C)             |
| `lhm_disk_load_percent`                  | Disk load (%)                     |
| `lhm_disk_level_percent`                 | Disk level (%)                    |
| `lhm_disk_factor_ratio`                  | Disk factor ratio (dimensionless) |
| `lhm_disk_data_bytes`                    | Disk data (bytes)                 |
| `lhm_disk_throughput_bytes_per_second`   | Disk throughput (bytes/s)         |
| `lhm_net_data_bytes`                     | Network data (bytes)              |
| `lhm_net_load_percent`                   | Network load (%)                  |
| `lhm_net_throughput_bytes_per_second`    | Network throughput (bytes/s)      |

## Notice / 注意事项

The `lhm_disk_data_bytes` metric may be inaccurate when the target device has encrypted drives.

`lhm_disk_data_bytes` 指标在目标设备存在加密驱动器时可能读取不准确。

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](.github/SECURITY.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
