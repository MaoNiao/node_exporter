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

// +build !noContainer_fs
package collector

import (
	"context"
	"bytes"
	"fmt"
	"strconv"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/common/log"
)

type ContainerFS struct {
	MountPoint string
	ContainerImage string
	ContainerName string
	ContainerId string
	Labels map[string]string
}

func GetAllContainerFS() (cfs []ContainerFS, err error) {
	dockerCli, err := client.NewEnvClient()
	if err != nil {
		log.Debugf("docker client start failed: %s", err)
		return
	}
	//get containers list
	containers, err := dockerCli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		containerJson, err := dockerCli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			log.Debugf("%q container get failed: %s", container.ID,err)
			continue
		}
		//get container labels (include pid, containers labels innotations)
		containerlabels := containerJson.Config.Labels

		cfs = append(cfs, ContainerFS{
			MountPoint:     fmt.Sprintf("/proc/%d/root", containerJson.State.Pid),
			ContainerImage: containerJson.Image,
			ContainerId:    containerJson.ID,
			ContainerName:  containerJson.Name,
			Labels:	        containerlabels,
		})
	}
	return cfs, nil
}