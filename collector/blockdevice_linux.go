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
	"syscall"
	"github.com/prometheus/common/log"
	//"github.com/kitt1987/superblock/pkg/xfs"
)

// GetStats returns blockdevice stats.
func (c *blockdeviceCollector) GetBlockDeviceStats() ([]blockdeviceStats, error) {
	//rbdd, err := readBlockDeviceDir()
	// ccfs, err := GetAllContainerFS()
	// if err != nil {
	// 	return nil, err
	// }
	//添加互斥锁，同一时间只有一用户对containerFsList进行读写
	stuckCFS.Lock()
	defer stuckCFS.Unlock()
	log.Infof("%d container fs will be exported", len(containerFsList))
	stats := []blockdeviceStats{}
	for _, containerfs := range containerFsList {
		buf := new(syscall.Statfs_t)
		err := syscall.Statfs(rootfsFilePath(containerfs.MountPoint), buf)
		if err != nil {
			log.Infof("Error on statfs() system call for %q: %s", rootfsFilePath(containerfs.MountPoint), err)
			continue
		}

		stats = append(stats, blockdeviceStats{
			labels:    blockdeviceLabels{
				podName:        containerfs.Labels["io.kubernetes.pod.name"],
				namespace:      containerfs.Labels["io.kubernetes.pod.namespace"],
				containerId:    containerfs.ContainerId,
				containerName:  containerfs.ContainerName,
				containerImage: containerfs.ContainerImage,
				pid:            containerfs.MountPoint,
			},
			size:      float64(buf.Blocks) * float64(buf.Bsize),
			free:      float64(buf.Bfree) * float64(buf.Bsize),
			avail:     float64(buf.Bavail) * float64(buf.Bsize),
		})
	}
	return stats, nil
}


//原方案，获取/sys/block/{device}/dm/name设备名，/dev/{device}从superblock的钱512字节中获取totalSize和avialSiza
// func readBlockDeviceDir() ([]blockdeviceLabels, error) {
// 	devlist, err := ioutil.ReadDir("/sys/block")
// 	if err != nil {
// 		log.Debugf("/sys/block read failed.  %s", err)
// 		return nil, err
// 	}

// 	blockdevices := []blockdeviceLabels{}
// 	for _, devfile := range devlist {
// 		var devname string = devfile.Name()
// 		//每次声明是不是不合适？
// 		//获取dm设备名
// 		if strings.Contains(devfile.Name(), "dm") {
// 			var buf bytes.Buffer
// 			buf.WriteString("/sys/block/")
// 			buf.WriteString(devfile.Name())
// 			buf.WriteString("/dm/name")
// 			devname, err = string(ioutil.ReadFile(buf.String())[:])
// 			if err != nil {
// 				log.Debugf("dm device name read failed :%q. %s", devname,err)
// 				continue
// 			}
// 		}

// 		var sbbuf bytes.Buffer
// 		sbbuf.WriteString("/dev/")
// 		sbbuf.WriteString(devfile.Name())
// 		sbdev := sbbuf.String()
// 		sb, err := xfs.GetSuperBlock(sbdev)
// 		if err != nil {
// 			log.Debugf("/dev path, device ID read failed :%q. %s", sb, err)
// 			continue
// 		}

// 		blockdevices = append(blockdevices, blockdeviceLabels{
// 			deviceId:   sbdev,
// 			deviceName: devname,
// 			totalSize:  uint64(sb.SB_blocksize) * uint64(sb.SB_dblocks),
// 			availSize:  uint64(sb.SB_blocksize) * uint64(sb.SB_fdblocks),
// 		})
// 	}
// 	return blockdevices, nil
// }
