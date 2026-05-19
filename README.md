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
## build the image with default registry, version and arch
$ make image
...
Successfully built f8e16ea6adb6
Successfully tagged registry.iluvatar.com.cn/k8s/ix-exporter:latest-x86_64

## build the image with customize registry, version and arch
$ REGISTRY=registry.iluvatar.com.cn/k8s VERSION=4.4.0 ARCH=x86_64 make image
# same effect: make image REGISTRY=... VERSION=... ARCH=...
# verify tag before building: REGISTRY=... VERSION=... ARCH=... make print-image
...
Successfully built f8e16ea6adb6
Successfully tagged registry.iluvatar.com.cn/k8s/ix-exporter:4.4.0-x86_64
```

If variables seem ignored, avoid `sudo make` (it drops your environment); use plain `make`, or `sudo -E make`, or pass variables on the `make` command line, for example `make image REGISTRY=... VERSION=... ARCH=...`.

## Usage

```shell
$ ./ix-exporter --help
NAME:
   IX Exporter - Generates Iluvatar coreX metrics in the prometheus format

USAGE:
   IX Exporter [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level value, -v value             Log level, 0-panic, 1-fatal, 2-error, 3-warn, 4-info, 5-debug, 6-trace. (default: 4) [$IX_EXPORTER_LOGLEVEL]
   --log-file value, -f value              Path of log file. (default: "/var/log/iluvatarcorex/ix-exporter/ix-exporter.log") [$IX_EXPORTER_LOGFILE]
   --enable-kubernetes, -k                 Enable kubernetes. (default: false) [$IX_EXPORTER_ENABLE_KUBERNETES]
   --metrics-config value, -c value        Path of metrics config file which contains of all fields. (default: "/opt/ix-exporter/metrics.yaml") [$IX_EXPORTER_METRICS_CONFIG]
   --remote-ix-hostengine value, -r value  Connect to remote ix-hostengine at <HOST>:<PORT>. (e.g. localhost:5777) [$IX_REMOTE_HOSTENGINE_INFO]
   --ip value                              Service IP. (default: "0.0.0.0") [$IX_EXPORTER_SERVICE_IP]
   --port value, -p value                  Service port. (default: "32021") [$IX_EXPORTER_SERVICE_PORT]
   --resource-name value                   Resource name of gpu in kubernetes. (default: "iluvatar.com/gpu") [$IX_EXPORTER_RESOURCE_NAME]
   --help, -h                              show help
```

Before running the **ix-exporter**, there are following preperations,

1. ensure that **Corex** was installed.
2. configure your [metrics.yaml](./etc/metrics.yaml) to enable metrics.
3. the **ix-exporter** use IxDCGM with embedded mode defaultly, if you want to connect to a remote ix-hostengine, please use `-r` option.

## Simple test of binary

```shell
./build/ix-exporter -c ./etc/metrics.yaml -p 32021
```

Default listening in `http://localhost:32021`.  

```shell
curl http://localhost:32021/metrics
```

## Quickstart on Kubernetes

See [Deploy IX Exporter in Kubernetes Cluster](deployment/README.md)

## Supported metrics and labels

See [metrics.yaml](etc/metrics.yaml)

## Example of gathering metrics on a GPU node

```shell
$ curl http://localhost:32021/metrics | grep ix_gpu_utilization
# HELP ix_gpu_utilization Utilization of iluvatar GPU (%).--:-- --:--:--  639k
1# TYPE ix_gpu_utilization gauge
0ix_gpu_utilization{driver="4.4.0",gpu="0",ixml="4.4.0",name="Iluvatar BI-V150S",node_name="infra-92",serial="24120026944896",uuid="GPU-6d2ec5fa-f293-57a3-9f2c-335f78120578"} 0
0ix_gpu_utilization{driver="4.4.0",gpu="1",ixml="4.4.0",name="Iluvatar BI-V150S",node_name="infra-92",serial="24120026944896",uuid="GPU-7edb0dc9-9291-5e13-9e1c-ad92672bdfec"} 0
 ix_gpu_utilization{driver="4.4.0",gpu="2",ixml="4.4.0",name="Iluvatar BI-V150S",node_name="infra-92",serial="00000000000000",uuid="GPU-a0ed6e8e-867a-5bb0-82a6-3697b7a3308d"} 0
1ix_gpu_utilization{driver="4.4.0",gpu="3",ixml="4.4.0",name="Iluvatar BI-V150S",node_name="infra-92",serial="00000000000000",uuid="GPU-9a5862a6-4db5-4a62-8096-4e23d4af6967"} 0
```

***Note***: if some metric values of gpu are not listed, it might be due to that some gpus not support a part of metrics.