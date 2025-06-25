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

package gpumanager

import (
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/utils/cpuset"

	"volcano.sh/volcano/pkg/scheduler/api"
)

type gpuAccumulator struct {
	topo          *api.GPUTopology
	details       api.GPUDetails
	numGPUsNeeded int
	result        cpuset.CPUSet
}

func newGPUAccumulator(topo *api.GPUTopology, availableGPUs cpuset.CPUSet, numGPUs int) *gpuAccumulator {
	return &gpuAccumulator{
		topo:          topo,
		details:       topo.GPUDetails.KeepOnly(availableGPUs),
		numGPUsNeeded: numGPUs,
		result:        cpuset.New(),
	}
}

func (a *gpuAccumulator) take(gpus cpuset.CPUSet) {
	a.result = a.result.Union(gpus)
	a.details = a.details.KeepOnly(a.details.GPUs().Difference(a.result))
	a.numGPUsNeeded -= gpus.Size()
}

func (a *gpuAccumulator) needs(n int) bool {
	return a.numGPUsNeeded >= n
}

func (a *gpuAccumulator) isSatisfied() bool {
	return a.numGPUsNeeded < 1
}

func (a *gpuAccumulator) isFailed() bool {
	return a.numGPUsNeeded > len(a.details)
}

// takeByTopology return the assigned gpuset
func takeByTopology(topo *api.GPUTopology, availableGPUs cpuset.CPUSet, numGPUs int) (cpuset.CPUSet, error) {
	acc := newGPUAccumulator(topo, availableGPUs, numGPUs)
	if acc.isSatisfied() {
		return acc.result, nil
	}
	if acc.isFailed() {
		return cpuset.New(), fmt.Errorf("not enough cpus available to satisfy request")
	}

	for _, c := range acc.details.GPUs().List() {
		klog.V(4).Infof("[gpumanager] takeByTopology: claiming GPU [%d]", c)
		if acc.needs(1) {
			acc.take(cpuset.New(c))
		}
		if acc.isSatisfied() {
			return acc.result, nil
		}
	}

	return cpuset.CPUSet{}, fmt.Errorf("not enough GPUs available to satisfy request: need %d", acc.numGPUsNeeded)
}
