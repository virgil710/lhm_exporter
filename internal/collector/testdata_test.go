package collector

// testSampleJSON contains a representative LHM JSON response for testing.
// This data mirrors the structure returned by the LibreHardwareMonitor web API.
const testSampleJSON = `{
  "id": 0,
  "Text": "Sensor",
  "Min": "Min",
  "Value": "Value",
  "Max": "Max",
  "ImageURL": "",
  "Children": [
    {
      "id": 1,
      "Text": "MSI-137K-DR4-23",
      "Min": "",
      "Value": "",
      "Max": "",
      "ImageURL": "images_icon/computer.png",
      "Children": [
        {
          "id": 2,
          "Text": "MSI PRO Z790-P WIFI DDR4 (MS-7E06)",
          "Min": "",
          "Value": "",
          "Max": "",
          "HardwareId": "/motherboard",
          "ImageURL": "images_icon/mainboard.png",
          "Children": [
            {
              "id": 3,
              "Text": "Nuvoton NCT6687D",
              "Min": "",
              "Value": "",
              "Max": "",
              "HardwareId": "/lpc/nct6687d/0",
              "ImageURL": "images_icon/chip.png",
              "Children": [
                {
                  "id": 4,
                  "Text": "Voltages",
                  "Min": "",
                  "Value": "",
                  "Max": "",
                  "ImageURL": "images_icon/voltage.png",
                  "Children": [
                    {"id": 5, "Text": "+12V", "Min": "12.072 V", "Value": "12.096 V", "Max": "12.120 V", "SensorId": "/lpc/nct6687d/0/voltage/0", "Type": "Voltage", "ImageURL": "", "Children": null},
                    {"id": 6, "Text": "Vcore", "Min": "0.752 V", "Value": "1.358 V", "Max": "1.418 V", "SensorId": "/lpc/nct6687d/0/voltage/2", "Type": "Voltage", "ImageURL": "", "Children": null}
                  ]
                },
                {
                  "id": 7,
                  "Text": "Temperatures",
                  "Min": "",
                  "Value": "",
                  "Max": "",
                  "ImageURL": "",
                  "Children": [
                    {"id": 8, "Text": "Motherboard", "Min": "28.0 °C", "Value": "32.0 °C", "Max": "45.0 °C", "SensorId": "/lpc/nct6687d/0/temperature/0", "Type": "Temperature", "ImageURL": "", "Children": null}
                  ]
                },
                {
                  "id": 9,
                  "Text": "Fans",
                  "Min": "",
                  "Value": "",
                  "Max": "",
                  "ImageURL": "",
                  "Children": [
                    {"id": 10, "Text": "Fan #1", "Min": "600 RPM", "Value": "1200 RPM", "Max": "1800 RPM", "SensorId": "/lpc/nct6687d/0/fan/0", "Type": "Fan", "ImageURL": "", "Children": null}
                  ]
                }
              ]
            }
          ]
        },
        {
          "id": 11,
          "Text": "Intel Core i7-13700K",
          "Min": "",
          "Value": "",
          "Max": "",
          "HardwareId": "/cpu",
          "ImageURL": "images_icon/cpu.png",
          "Children": [
            {
              "id": 12,
              "Text": "Temperatures",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 13, "Text": "Package", "Min": "25.0 °C", "Value": "55.0 °C", "Max": "95.0 °C", "SensorId": "/cpu/0/temperature/0", "Type": "Temperature", "ImageURL": "", "Children": null}
              ]
            },
            {
              "id": 14,
              "Text": "Load",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 15, "Text": "CPU Total", "Min": "0.0 %", "Value": "25.5 %", "Max": "100.0 %", "SensorId": "/cpu/0/load/0", "Type": "Load", "ImageURL": "", "Children": null}
              ]
            },
            {
              "id": 16,
              "Text": "Clocks",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 17, "Text": "CPU Core #1", "Min": "800 MHz", "Value": "3600 MHz", "Max": "5300 MHz", "SensorId": "/cpu/0/clock/0", "Type": "Clock", "ImageURL": "", "Children": null}
              ]
            },
            {
              "id": 18,
              "Text": "Power",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 19, "Text": "Package", "Min": "10.0 W", "Value": "65.0 W", "Max": "253.0 W", "SensorId": "/cpu/0/power/0", "Type": "Power", "ImageURL": "", "Children": null}
              ]
            }
          ]
        },
        {
          "id": 20,
          "Text": "Generic Memory",
          "Min": "",
          "Value": "",
          "Max": "",
          "HardwareId": "/ram",
          "ImageURL": "images_icon/ram.png",
          "Children": [
            {
              "id": 21,
              "Text": "Load",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 22, "Text": "Memory", "Min": "4.0 GB", "Value": "16.2 GB", "Max": "32.0 GB", "SensorId": "/ram/0/load/0", "Type": "Load", "ImageURL": "", "Children": null}
              ]
            },
            {
              "id": 23,
              "Text": "Data",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 24, "Text": "Available Memory", "Min": "", "Value": "15.8 GB", "Max": "", "SensorId": "/ram/0/data/0", "Type": "Data", "ImageURL": "", "Children": null}
              ]
            }
          ]
        },
        {
          "id": 25,
          "Text": "NVIDIA GeForce RTX 4070",
          "Min": "",
          "Value": "",
          "Max": "",
          "HardwareId": "/gpu/nvidia/0",
          "ImageURL": "images_icon/gpu.png",
          "Children": [
            {
              "id": 26,
              "Text": "Temperatures",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 27, "Text": "GPU Core", "Min": "25.0 °C", "Value": "48.0 °C", "Max": "90.0 °C", "SensorId": "/gpu/nvidia/0/temperature/0", "Type": "Temperature", "ImageURL": "", "Children": null}
              ]
            },
            {
              "id": 28,
              "Text": "Load",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 29, "Text": "GPU Core", "Min": "0.0 %", "Value": "15.0 %", "Max": "100.0 %", "SensorId": "/gpu/nvidia/0/load/0", "Type": "Load", "ImageURL": "", "Children": null}
              ]
            },
            {
              "id": 30,
              "Text": "Fans",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 31, "Text": "Fan #1", "Min": "0 RPM", "Value": "900 RPM", "Max": "2000 RPM", "SensorId": "/gpu/nvidia/0/fan/0", "Type": "Fan", "ImageURL": "", "Children": null}
              ]
            }
          ]
        },
        {
          "id": 32,
          "Text": "Samsung SSD 990 PRO 2TB",
          "Min": "",
          "Value": "",
          "Max": "",
          "HardwareId": "/nvme/0",
          "ImageURL": "images_icon/drive.png",
          "Children": [
            {
              "id": 33,
              "Text": "Temperatures",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 34, "Text": "Temperature", "Min": "25.0 °C", "Value": "38.0 °C", "Max": "70.0 °C", "SensorId": "/nvme/0/temperature/0", "Type": "Temperature", "ImageURL": "", "Children": null}
              ]
            },
            {
              "id": 35,
              "Text": "Load",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 36, "Text": "Used Space", "Min": "", "Value": "45.2 %", "Max": "", "SensorId": "/nvme/0/load/0", "Type": "Load", "ImageURL": "", "Children": null}
              ]
            },
            {
              "id": 37,
              "Text": "Data",
              "Min": "",
              "Value": "",
              "Max": "",
              "ImageURL": "",
              "Children": [
                {"id": 38, "Text": "Read Rate", "Min": "", "Value": "150.0 MB/s", "Max": "", "SensorId": "/nvme/0/data/0", "Type": "Data", "ImageURL": "", "Children": null}
              ]
            }
          ]
        }
      ]
    }
  ]
}`
