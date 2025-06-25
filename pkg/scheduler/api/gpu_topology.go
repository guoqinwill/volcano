/*
Copyright 2025 The Volcano Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/utils/cpuset"

	nodeinfov1alpha1 "volcano.sh/apis/pkg/apis/nodeinfo/v1alpha1"
)

// GPUDetails is a map from GPU ID
type GPUDetails map[string]nodeinfov1alpha1.GPUInfo

// GPUTopology contains details of node gpu, where :
// NUMA Node - NUMA cell, cadvisor - Node
type GPUTopology struct {
	NumGPUs      int
	NumNUMANodes int
	GPUDetails   GPUDetails
}

// GPUNUMANodeID returns the NUMA node ID which the given GPU belongs to.
func (topo *GPUTopology) GPUNUMANodeID(gpu string) (int, error) {
	info, ok := topo.GPUDetails[gpu]
	if !ok {
		return -1, fmt.Errorf("unknown GPU ID: %s", gpu)
	}
	return info.NUMANodeID, nil
}

// KeepOnly returns a new GPUDetails object with only the supplied gpus.
func (d GPUDetails) KeepOnly(gpus cpuset.CPUSet) GPUDetails {
	result := GPUDetails{}
	for gpu, info := range d {
		if strings.HasPrefix(gpu, "nvidia") {
			idStr := strings.TrimPrefix(gpu, "nvidia")
			if id, err := strconv.Atoi(idStr); err == nil && gpus.Contains(id) {
				result[gpu] = info
			}
		}
	}
	return result
}

// NUMANodes returns all of the NUMANode IDs associated with the GPUs in this
// GPUDetails.
func (d GPUDetails) NUMANodes() cpuset.CPUSet {
	var numaNodeIDs []int
	for _, info := range d {
		numaNodeIDs = append(numaNodeIDs, info.NUMANodeID)
	}
	return cpuset.New(numaNodeIDs...)
}

// GPUs returns all of the logical GPU IDs in this GPUDetails.
func (d GPUDetails) GPUs() cpuset.CPUSet {
	var gpuIDs []int
	for gpuID := range d {
		if strings.HasPrefix(gpuID, "nvidia") {
			idStr := strings.TrimPrefix(gpuID, "nvidia")
			if id, err := strconv.Atoi(idStr); err == nil {
				gpuIDs = append(gpuIDs, id)
			}
		}
	}
	return cpuset.New(gpuIDs...)
}

// GPUsInNUMANodes returns all of the logical GPU IDs associated with the given
// NUMANode IDs in this GPUDetails.
func (d GPUDetails) GPUsInNUMANodes(ids ...int) cpuset.CPUSet {
	var gpuIDs []int
	for _, id := range ids {
		for gpu, info := range d {
			if info.NUMANodeID == id {
				if strings.HasPrefix(gpu, "nvidia") {
					idStr := strings.TrimPrefix(gpu, "nvidia")
					if gpuid, err := strconv.Atoi(idStr); err == nil {
						gpuIDs = append(gpuIDs, gpuid)
					}
				}
			}
		}
	}
	return cpuset.New(gpuIDs...)
}
