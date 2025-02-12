# IX-Exporter

IX-Exporter is a http server to expose Iluvatar GPU node information.

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
   ix-exporter - Generates Iluvatar coreX metrics in the prometheus format

USAGE:
   ix-exporter [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level value, -v value             Log level, 0-panic, 1-fatal, 2-error, 3-warn, 4-info, 5-debug, 6-trace. (default: 4) [$IX_EXPORTER_LOGLEVEL]
   --log-file value, -f value              Path of log file. (default: "/tmp/log/ix-exporter.log") [$IX_EXPORTER_LOGFILE]
   --enable-kubernetes, -k                 Enable kubernetes. (default: false) [$IX_EXPORTER_ENABLE_KUBERNETES]
   --metrics-config value, -c value        Path of metrics config file which contains of all fields. (default: "/etc/ixexporter/metrics.yaml") [$IX_EXPORTER_METRICS_CONFIG]
   --remote-ix-hostengine value, -r value  Connect to remote ix-hostengine at <HOST>:<PORT>. (e.g. 10.10.2.6:5777) [$IX_REMOTE_HOSTENGINE_INFO]
   --ip value                              Service IP. (default: "0.0.0.0") [$IX_EXPORTER_SERVICE_IP]
   --port value, -p value                  Service port. (default: "32021") [$IX_EXPORTER_SERVICE_PORT]
   --help, -h                              show help
```

Before running the **ix-exporter**, there are following preperations,

1. ensure that **Corex** was installed.  
2. configure your [metrics.yaml](./etc/metrics.yaml) to enable metrics.
3. the **ix-exporter** use IxDCGM with embedded mode defaultly, if you want to connect to a remote ix-hostengine, please use `-r` option.


## Simple test of binary
```shell
$ ./build/ix-exporter -c ./etc/metrics.yaml -r localhost:5777 -p 32021
```

Default listening in `http://localhost:32021`.  
```shell
$ curl http://localhost:32021/metrics
```

# Quickstart on Kubernetes
## Deploy IX Exporter in Kubernetes Cluster
The next step is to run ix-expoter on each node as a Daemonset.

```bash
$ sudo kubectl apply -f deployment/static/ix-exporter.yaml
```

Check if the ```ix-exporter``` daemonset is deployed successfully in kubernetes cluster:
```bash
$ sudo kubectl -n kube-system get ds ix-exporter
NAME          DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE
ix-exporter   1         1         1       1            1            
```

## Config Prometheus
If the Prometheus is deployed in kubernetes cluster, the next step is to create the ```ix-expoter``` ServiceMonitor object to integrate with Prometheus:  
```bash
$ sudo kubectl -n monitoring apply -f deployment/static/service-monitor.yaml
```
**Note**: make sure your prometheus is deployed in ```monitoring``` namespace.

Check if the ```ix-expoter``` ServiceMonitor object is created successfully in kubernetes cluster:
```bash
$ sudo kubectl -n monitoring get servicemonitor ix-exporter
NAME          AGE
ix-exporter   8s
```

It may take a few minutes for ix-exporter to start publishing the metrics to Prometheus. The metrics availability can be verified by typing ix metric key (e.g. **ix_gpu_utilization**) in the event bar to determine if the GPU metrics are visible.

# Example of gathering metrics on a GPU node
```shell
$ curl http://localhost:32021/metrics
# HELP ix_ecc_dbe_vol_status The double-bit volatile ecc errors status. if the value is 1, errors occurred, otherwise, no errors.
# TYPE ix_ecc_dbe_vol_status gauge
ix_ecc_dbe_vol_status{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_ecc_dbe_vol_status{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 0
ix_ecc_dbe_vol_status{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
ix_ecc_dbe_vol_status{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 0
# HELP ix_ecc_sbe_vol_status The single-bit volatile ecc errors status. if the value is 1, errors occurred, otherwise, no errors.
# TYPE ix_ecc_sbe_vol_status gauge
ix_ecc_sbe_vol_status{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_ecc_sbe_vol_status{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 0
ix_ecc_sbe_vol_status{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
ix_ecc_sbe_vol_status{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 0
# HELP ix_gpu_utilization The utilization of iluvatar GPU (%).
# TYPE ix_gpu_utilization gauge
ix_gpu_utilization{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_gpu_utilization{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 0
ix_gpu_utilization{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 100
ix_gpu_utilization{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 100
# HELP ix_mem_clock Mem clock of iluvatar GPU (MHz).
# TYPE ix_mem_clock gauge
ix_mem_clock{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 1200
ix_mem_clock{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 1600
ix_mem_clock{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 1600
ix_mem_clock{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 1600
# HELP ix_mem_free The free physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_free gauge
ix_mem_free{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 32511
ix_mem_free{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 32652
ix_mem_free{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 31870
ix_mem_free{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 31870
# HELP ix_mem_total The total physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_total gauge
ix_mem_total{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 32768
ix_mem_total{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 32768
ix_mem_total{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 32768
ix_mem_total{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 32768
# HELP ix_mem_used The used physical memory of iluvatar GPU (MiB).
# TYPE ix_mem_used gauge
ix_mem_used{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 257
ix_mem_used{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 116
ix_mem_used{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 898
ix_mem_used{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 898
# HELP ix_mem_utilization The memory utilization of iluvatar GPU (%).
# TYPE ix_mem_utilization gauge
ix_mem_utilization{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 1
ix_mem_utilization{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 1
ix_mem_utilization{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 3
ix_mem_utilization{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 3
# HELP ix_pcie_replay_counter The PCIe replay counter.
# TYPE ix_pcie_replay_counter gauge
ix_pcie_replay_counter{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_pcie_replay_counter{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 0
ix_pcie_replay_counter{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
ix_pcie_replay_counter{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 0
# HELP ix_pcie_rx_throughput The PCIe rx (read) data including both header and payload (KB/s).
# TYPE ix_pcie_rx_throughput gauge
ix_pcie_rx_throughput{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_pcie_rx_throughput{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 28
ix_pcie_rx_throughput{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 77433
ix_pcie_rx_throughput{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 74598
# HELP ix_pcie_tx_throughput The PCIe tx (transmit) data including both header and payload (KB/s).
# TYPE ix_pcie_tx_throughput gauge
ix_pcie_tx_throughput{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_pcie_tx_throughput{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 24180
ix_pcie_tx_throughput{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
ix_pcie_tx_throughput{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 0
# HELP ix_power_usage The power usage of iluvatar GPU.
# TYPE ix_power_usage gauge
ix_power_usage{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 55
ix_power_usage{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 132
ix_power_usage{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 206
ix_power_usage{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 210
# HELP ix_process_info The process info of iluvatar GPU (MiB).
# TYPE ix_process_info gauge
ix_process_info{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",process_name="",process_pid="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_process_info{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",process_name="",process_pid="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 0
ix_process_info{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",process_name="./gemm_perf --i 2,3 --d 0 --m 1024 --l 1000 ",process_pid="49685",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 782
ix_process_info{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",process_name="./gemm_perf --i 2,3 --d 0 --m 1024 --l 1000 ",process_pid="49685",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 782
# HELP ix_sm_clock Sm clock of iluvatar GPU (MHz).
# TYPE ix_sm_clock gauge
ix_sm_clock{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 1500
ix_sm_clock{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 1500
ix_sm_clock{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 1600
ix_sm_clock{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 1625
# HELP ix_sm_utilization The utilization of SM (%).
# TYPE ix_sm_utilization gauge
ix_sm_utilization{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 0
ix_sm_utilization{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 1
ix_sm_utilization{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 1
# HELP ix_temperature The temperature of the iluvatar GPU(C).
# TYPE ix_temperature gauge
ix_temperature{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 35
ix_temperature{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 57
ix_temperature{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 60
ix_temperature{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 59
# HELP ix_xid_errors The Value of the last xid error encountered.
# TYPE ix_xid_errors gauge
ix_xid_errors{container="",gpu="0",name="Iluvatar BI-V100",namespace="",node_name="infra-92",pod="",uuid="GPU-4a8348cb-505c-507f-8df7-ff3c796e3033"} 0
ix_xid_errors{container="",gpu="1",name="Iluvatar MR-V50",namespace="",node_name="infra-92",pod="",uuid="GPU-2421fa19-18cf-47bb-b629-1ae3e642436d"} 0
ix_xid_errors{container="",gpu="2",name="Iluvatar BI-V150S",namespace="",node_name="infra-92",pod="",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
ix_xid_errors{container="busybox",gpu="3",name="Iluvatar BI-V150S",namespace="default",node_name="infra-92",pod="test-pod",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 0
```

***Note***: if some metric values of gpu are not listed, it might be due to that some gpus not support a part of metrics.