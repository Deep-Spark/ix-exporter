package reset

import (
	"os"
	"time"
)

const GpuResetCMPrefix = "ix-gpu-reset-"
const GpuResetKey = "gpuReset"
const ExporterOccupyKey = "ix-exporter"

type ResetEvent int

const ResetToFalse ResetEvent = 0
const ResetToTrue ResetEvent = 1

type PatchOperation int

const (
	PatchOpDeleteKey PatchOperation = iota
	PatchOpSetTrue
	PatchOpSetFalse
)

const RetryCount = 3
const RetryWaitTime = 1000 * time.Millisecond

var ShutdownFlag = make(chan struct{})

type GpuResetInfo struct {
	NodeName string          `json:"nodename"             yaml:"nodename"`
	Reset    bool            `json:"reset"                yaml:"reset"`
	Occupy   map[string]bool `json:"occupy"               yaml:"occupy"`
}

type GpuResetInformer interface {
	Start()
	Stop()

	AddCm(obj interface{}, sig chan os.Signal, ch chan ResetEvent)

	// When the reset change to true, push SIGHUP to sig and push ResetToTrue to ch.
	// When the reset change to false, push ResetToFalse to ch.
	UpdateCm(obj interface{}, sig chan os.Signal, ch chan ResetEvent)

	DeleteCm(sig chan os.Signal, ch chan ResetEvent)

	// Patch the configmap with the given patch operation.
	PatchCm(op PatchOperation) error
}
