// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !noblockdevice

package collector

import (
	"github.com/prometheus/common/log"
	"sync"
	"time"
	//"regexp"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Arch-dependent implementation must define:
// * defIgnoredMountPoints
// * defIgnoredFSTypes
// * filesystemLabelNames
// * filesystemCollector.GetStats

var (
	// ignoredMountPoints = kingpin.Flag(
	// 	"collector.blockdevice.ignored-mount-points",
	// 	"Regexp of mount points to ignore for blockdevice collector.",
	// ).Default(defIgnoredMountPoints).String()
	// ignoredFSTypes = kingpin.Flag(
	// 	"collector.blockdevice.ignored-fs-types",
	// 	"Regexp of filesystem types to ignore for blockdevice collector.",
	// ).Default(defIgnoredFSTypes).String()

	hostProcPath = kingpin.Flag("collector.procpath","mount the host directory, default /proc.").Default("/proc").String()
	collectorTime = kingpin.Flag("collector.containersTime","the interval for collecting containerList.").Default("10").Int64()
	blockdeviceLabelNames = []string{"podName", "namespace", "containerId", "containerName", "containerImage", "pid"}
)

type blockdeviceCollector struct {
	// ignoredMountPointsPattern     *regexp.Regexp
	// ignoredFSTypesPattern         *regexp.Regexp
	sizeDesc, freeDesc, availDesc *prometheus.Desc
}

type blockdeviceLabels struct {
	//deviceId, deviceName, totalSize, availSize string
	podName, namespace, containerId, containerName, containerImage, pid string
}

type blockdeviceStats struct {
	labels            blockdeviceLabels
	size, free, avail       float64
}

var (
	containerFsList               []ContainerFS
	stuckCFS                      sync.Mutex
)
//init时开启多线程收集容器信息，并注册blockdeviceCollector
func init() {
	go func(){
		t := time.NewTimer(0)
		for {
			select {
			case <-t.C:
				gacfs, err := GetAllContainerFS()
				if err != nil {
					panic(err)
				} else {
					//添加互斥锁，同一时间只有一用户对containerFsList进行读写
					stuckCFS.Lock()
					containerFsList = gacfs
					stuckCFS.Unlock()
					log.Infof("%d container fs loaded", len(gacfs))
				}
				kingpin.Parse()
				t.Reset(time.Second * time.Duration(*collectorTime))
			}
		}
	}()

	registerCollector("blockdevice", defaultEnabled, NewBlockdeviceCollector)
}

// NewBlockdeviceCollector returns a new Collector exposing blockdevice stats.
func NewBlockdeviceCollector() (Collector, error) {
	c := &blockdeviceCollector{}

	subsystem := "blockdevice"
	// mountPointPattern := regexp.MustCompile(*ignoredMountPoints)
	// filesystemsTypesPattern := regexp.MustCompile(*ignoredFSTypes)

	c.sizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "size_bytes"),
		"Filesystem size in bytes.",
		blockdeviceLabelNames, nil,
	)

	c.freeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "free_bytes"),
		"Filesystem free space in bytes.",
		blockdeviceLabelNames, nil,
	)

	c.availDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "avail_bytes"),
		"Filesystem space available to non-root users in bytes.",
		blockdeviceLabelNames, nil,
	)

	return c, nil
}

func (c *blockdeviceCollector) Update(ch chan<- prometheus.Metric) error {
	stats, err := c.GetBlockDeviceStats()
	if err != nil {
		return err
	}
	// Make sure we expose a metric once, even if there are multiple mounts
	seen := map[blockdeviceLabels]bool{}
	for _, s := range stats {
		if seen[s.labels] {
			continue
		}
		seen[s.labels] = true

		//ch <- prometheus.MustNewConstMetric(
		//	c.deviceErrorDesc, prometheus.GaugeValue,
		// 	s.deviceError, s.labels.device, s.labels.mountPoint, s.labels.fsType,
		// )
		// if s.deviceError > 0 {
		// 	continue
		// }

		ch <- prometheus.MustNewConstMetric(
			c.sizeDesc, prometheus.GaugeValue,
			s.size, s.labels.podName, s.labels.namespace, s.labels.containerId, s.labels.containerName, s.labels.containerImage, s.labels.pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.freeDesc, prometheus.GaugeValue,
			s.free, s.labels.podName, s.labels.namespace, s.labels.containerId, s.labels.containerName, s.labels.containerImage, s.labels.pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availDesc, prometheus.GaugeValue,
			s.avail, s.labels.podName, s.labels.namespace, s.labels.containerId, s.labels.containerName, s.labels.containerImage, s.labels.pid,
		)

	}
	return nil
}
