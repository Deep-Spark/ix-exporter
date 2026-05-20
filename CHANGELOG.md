# IX-Exporter Changelog

## v4.4.0

- Introduced new metrics: `ix_gpu_temperature`, `ix_mem_temperature`, `ix_xid_errors`
- Added new labels: `driver` (GPU driver version), `ixml` (ixml version), `serial`(GPU serial number)
- Synchronized configuration from the `ix-config` ConfigMap using an informer
- Enabled GPU reset and hot-plug functionality
- Resolved various security vulnerabilities and bugs

## v4.3.0

- Added support for deployment with Helm Chart
- Simultaneous use of `go-ixml` and `go-ixdcgm` interfaces
- Retrieved running process information via `go-ixdcgm`
- Fixed issues related to metric retrieval

## v4.2.0

- Expanded metrics and labels
- Integrated support for Kubernetes
- Utilized `go-ixml` interface

## v4.1.2

- Initial release
