# IX-Expoter

ix-expoter is a http server to expose Iluvatar GPU node information.

## Build binary and image

Build the executable binary `ix-exporter` to the `build` directory.
```shell
$ make build
$ ls build/ix-exporter
build/ix-exporter
```

Build the image
```shell
## build the image with default registry and version
$ make image
...
Successfully built f8e16ea6adb6
Successfully tagged ix-exporter:4.2.0-x86_64

## build the image with customize registry and version 
$ REGISTRY=iluvatar.com/release VERSION=v4.2.1 make image
...
Successfully built f8e16ea6adb6
Successfully tagged iluvatar.com/release/ix-exporter:v4.2.1-x86_64
```

## Usage

```shell
$ ./ix-exporter --help
NAME:
   ix-exporter - Export iluvatar data to Prometheus

USAGE:
   ix-exporter [global options] command [command options]

GLOBAL OPTIONS:
   --log-level value, -l value       Log level, 0-debug, 1-info, 2-warning, 3-error, 4-fatal(default 0) (default: 0) [$IX_EXPORTER_LOGLEVEL]
   --log-file value, -f value        Log file path name. (default: "/tmp/log/ix-exporter.log") [$IX_EXPORTER_LOGFILE]
   --enable-kubernetes, -k           Enable Kubernetes mode. (default: true) [$IX_EXPORTER_ENABLE_KUBERNETES]
   --metrics-config value, -c value  Metrics config file which contains of all fields. (default: "/etc/ixexporter/metrics.yaml") [$IX_EXPORTER_METRICS_CONFIG]
   --ip value                        Service IP. (default: "0.0.0.0") [$IX_EXPORTER_SERVICE_IP]
   --port value, -p value            Service port (default: "32021") [$IX_EXPORTER_SERVICE_PORT]
   --help, -h                        show help
```

Before running the **ix-exporter**, there are following preperations,

1. ensure that **Corex** was installed.

2. configure your [exporter.yaml](./etc/exporter.yaml) to enable metrics.

```shell
$ sudo ./ix-exporter -c /path/to/your/exporter.yaml
```

Default listening in `http://localhost:32021`. 

## Config Prometheus and Grafana
- You should copy **gpu-iluvatar job** in `prometheus_config_sample.yml` to your Prometheus config file(default location:/etc/prometheus/prometheus.yml). Then, you need to update your prometheus service. 

- After starting Grafa, you should import `IX-exporter-dashboard.json` in grafana directory to create IX-exporter dashboard. 

## Example output

```shell
$ curl http://localhost:32021/metrics
# HELP ix_fan_speed Fan speed of iluvatar GPU.
# TYPE ix_fan_speed gauge
ix_fan_speed{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_fan_speed{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 0
ix_fan_speed{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
# HELP ix_gpu_utilization The utilization of iluvatar GPU (%).
# TYPE ix_gpu_utilization gauge
ix_gpu_utilization{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_gpu_utilization{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 0
ix_gpu_utilization{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
# HELP ix_mem_clock Mem clock of iluvatar GPU (MHz).
# TYPE ix_mem_clock gauge
ix_mem_clock{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 1200
ix_mem_clock{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 1600
ix_mem_clock{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 1600
# HELP ix_mem_free The free physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_free gauge
ix_mem_free{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 32511
ix_mem_free{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 32652
ix_mem_free{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 32652
# HELP ix_mem_total The total physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_total gauge
ix_mem_total{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 32768
ix_mem_total{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 32768
ix_mem_total{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 32768
# HELP ix_mem_used The used physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_used gauge
ix_mem_used{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 257
ix_mem_used{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 116
ix_mem_used{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 116
# HELP ix_mem_utilization The memory utilization of iluvatar GPU (%).
# TYPE ix_mem_utilization gauge
ix_mem_utilization{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 1
ix_mem_utilization{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 1
ix_mem_utilization{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 1
# HELP ix_power_usage The power usage of iluvatar GPU.
# TYPE ix_power_usage gauge
ix_power_usage{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 54
ix_power_usage{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 131
ix_power_usage{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 13
# HELP ix_process_info The process info of iluvatar GPU (MiB).
# TYPE ix_process_info gauge
ix_process_info{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",process_name="",process_pid="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_process_info{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",process_name="",process_pid="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 0
ix_process_info{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",process_name="",process_pid="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
# HELP ix_sm_clock Sm clock of iluvatar GPU (MHz).
# TYPE ix_sm_clock gauge
ix_sm_clock{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 1500
ix_sm_clock{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 1500
ix_sm_clock{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 500
# HELP ix_sm_utilization The utilization of SM (%).
# TYPE ix_sm_utilization gauge
ix_sm_utilization{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 0
ix_sm_utilization{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
# HELP ix_temperature The temperature of the iluvatar GPU(C).
# TYPE ix_temperature gauge
ix_temperature{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 34
ix_temperature{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 57
ix_temperature{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 31
# HELP ix_xid_errors The Value of the last xid error encountered.
# TYPE ix_xid_errors gauge
ix_xid_errors{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_xid_errors{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="",pod="",uuid="GPU-50351a81-6f42-4746-9981-6e4401848ba5"} 0
ix_xid_errors{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
```
