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
	"io/ioutil"
	"bufio"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"ioutil"
	"github.com/prometheus/common/log"
	"github.com/kitt1987/superblock/pkg/xfs"
)

const (
	defIgnoredMountPoints = "^/(dev|proc|sys|var/lib/docker/.+)($|/)"
	defIgnoredFSTypes     = "^(autofs|binfmt_misc|cgroup|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|sysfs|tracefs)$"
	readOnly              = 0x1 // ST_RDONLY
	mountTimeout          = 30 * time.Second
)

// GetStats returns blockdevice stats.
func (c *blockdeviceCollector) GetBlockDeviceStats() ([]blockdeviceStats, error) {
	rbdd, err := readBlockDeviceDir()
	if err != nil {
		return nil, err
	}
	stats := []blockdeviceStats{}
	for _, labels := range rbdd {
		buf := new(syscall.Statfs_t)
		err = syscall.Statfs(rootfsFilePath(labels.deviceId), buf)
		if err != nil {
			log.Debugf("Error on statfs() system call for %q: %s", rootfsFilePath(labels.deviceId), err)
			continue
		}

		stats = append(stats, blockdeviceStats{
			labels:    labels,
			size:      float64(buf.Blocks) * float64(buf.Bsize),
			avail:     float64(buf.Bavail) * float64(buf.Bsize),
		})
	}
	return stats, nil
}

func readBlockDeviceDir() ([]blockdeviceLabels, error) {
	devlist, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		log.Debugf("/sys/block read failed.  %s", err)
		return nil, err
	}
	defer devlist.Close()

	blockdevices := []blockdeviceLabels{}
	for _, devfile := range devlist {
		var devname string = devfile.Name()
		//每次声明是不是不合适？
		//获取dm设备名
		if strings.Contains(devfile.Name(), "dm") {
			var buf bytes.Buffer
			buf.WriteString("/sys/block/")
			buf.WriteString(devfile.Name())
			buf.WriteString("/dm/name")
			devname, err = ioutil.ReadFile(buf.String()) 
			if err != nil {
				log.Debugf("dm device name read failed :%q. %s", devname,err)
				continue
			}
		}

		var sbbuf bytes.Buffer
		sbbuf.WriteString("/dev/")
		sbbuf.WriteString(devfile.Name())
		sbdev := sbbuf.String()
		sb, err := xfs.GetSuperBlock(sbdev)
		if err != nil {
			log.Debugf("/dev path, device ID read failed :%q. %s", sb, err)
			continue
		}

		blockdevices = append(blockdevices, blockdeviceLabels{
			deviceId:   sbdev,
			deviceName: devname,
			totalSize:  uint64(sb.SB_blocksize) * uint64(sb.SB_dblocks),
			availSize:  uint64(sb.SB_blocksize) * uint64(sb.SB_fdblocks),
		})
	}
	return blockdevices, nil
}
