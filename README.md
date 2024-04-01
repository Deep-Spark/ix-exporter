# IX-Exporter

ix-exporter is an HTTP server that exposes Iluvatar GPU node information.

## Build

```shell
make
```

## Usage

```shell
./ix-exporter --help
Usage of ixexporter:
  -a,  string   Metrics config file which contains of all fields. (default "localhost")
  -c,  string   Metrics config file which contains of all fields. (default "/etc/ixexporter/exporter.yaml")
  -p,  uint     Service port. (default 32021)
  -r,  string   Metrics router. (default "/metrics")
```

Before running the **ix-exporter**, there are following preparations,

1. Ensure that **Corex** was installed.

2. Configure [exporter.yaml](./etc/exporter.yaml) to enable metrics.

```shell
sudo ./ix-exporter
```

Default Listening in `http://localhost:32021`. 

## Config Prometheus and Grafana

- You should copy **gpu-iluvatar job** in prometheus_config_sample.yml to your Prometheus config file(default location:/etc/prometheus/prometheus.yml). Then, you need to update your Prometheus service. 

- After starting Grafana, you should import **IX-exporter-dashboard.json** in grafana directory to create IX-exporter dashboard. 

## Example output

```shell
curl http://localhost:32021/metrics
# HELP ix_fan_speed Fan speed of iluvatar GPU
# TYPE ix_fan_speed gauge
ix_fan_speed{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 0
ix_fan_speed{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 0
# HELP ix_gpu_utilization The utilization of iluvatar GPU (%).
# TYPE ix_gpu_utilization gauge
ix_gpu_utilization{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 0
ix_gpu_utilization{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 0
# HELP ix_mem_clock Mem clock of iluvatar GPU (MHz).
# TYPE ix_mem_clock gauge
ix_mem_clock{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 1200
ix_mem_clock{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 1200
# HELP ix_mem_free The free physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_free gauge
ix_mem_free{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 32255
ix_mem_free{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 32255
# HELP ix_mem_total The total physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_total gauge
ix_mem_total{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 32768
ix_mem_total{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 32768
# HELP ix_mem_used The used physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_used gauge
ix_mem_used{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 513
ix_mem_used{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 513
# HELP ix_mem_utilization The memory utilization of iluvatar GPU (%).
# TYPE ix_mem_utilization gauge
ix_mem_utilization{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 2
ix_mem_utilization{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 2
# HELP ix_power_usage The power usage of iluvatar GPU.
# TYPE ix_power_usage gauge
ix_power_usage{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 35
ix_power_usage{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 35
# HELP ix_sm_clock Sm clock of iluvatar GPU (MHz).
# TYPE ix_sm_clock gauge
ix_sm_clock{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 1000
ix_sm_clock{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 1000
# HELP ix_temperature The temperature of iluvatar GPU (C).
# TYPE ix_temperature gauge
ix_temperature{gpu="0",name="AIP-BI",uuid="GPU-fd35cbcb-bb08-4fe5-a4f0-49633e4681f0"} 37
ix_temperature{gpu="1",name="AIP-BI",uuid="GPU-1fc82808-c51b-4fa0-979c-582a6e530166"} 38
```

## License

Copyright (c) 2024 Iluvatar CoreX. All rights reserved. This project has an Apache-2.0 license, as found in the [LICENSE](LICENSE) file.
