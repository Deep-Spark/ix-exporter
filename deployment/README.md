# Deploy IX Exporter in Kubernetes Cluster

## Deploy IX Exporter with Static Yaml Files

Deploy ix-expoter on each node as a Daemonset.  

```bash
sudo kubectl apply -f static/ix-exporter.yaml
```

If the Prometheus was deployed, you could optionally create a ServiceMonitor object to integrate ix-exporter with it.  

```bash
sudo kubectl -n monitoring apply -f static/service-monitor.yaml
```

**Note**: make sure your Prometheus was deployed in `monitoring` namespace.

## Deploy IX Exporter with Helm Chart

```bash
$ sudo helm install ix-exporter deployment/helm/ix-exporter \
    -n kube-system \ 
    --insecure-skip-tls-verify \
    --set image=<your_image_name> \ # Replace <your_image_name> with your image name
    --set daemonset.logLevel=4 \ # Set the log level (0-6), 0-panic, 1-fatal, 2-error, 3-warn, 4-info, 5-debug, 6-trace, defalut is 4
    --set daemonset.resourceName=<your_resource_name> \ # Set the resource name of gpu in k8s, default is "iluvatar.com/gpu"
    --set service.ip=<your_service_ip> \ # Set the service ip of metrics, default is 0.0.0.0
    --set service.port=<your_service_port> \ # Set the service port of metrics, default is 32021
    --set serviceMonitor.enabled=true # Enable ServiceMonitor object if you want to integrate ix-exporter with Prometheus, default is false
```

**Note**: typically, you only need to set the image.

## Check the status of IX Exporter

Check if the `ix-exporter` daemonset is deployed successfully in kubernetes cluster:

```bash
$ sudo kubectl -n kube-system get ds ix-exporter
NAME          DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE
ix-exporter   1         1         1       1            1            
```

Check if the `ix-exporter` ServiceMonitor object is created successfully in Kubernetes cluster:

```bash
$ sudo kubectl -n monitoring get servicemonitor ix-exporter
NAME          AGE
ix-exporter   8s
```

It may take a few minutes for ix-exporter to start publishing the metrics to Prometheus. The metrics availability can be verified by typing ix metric key (e.g. **ix_gpu_utilization**) in the event bar to determine if the GPU metrics are visible.