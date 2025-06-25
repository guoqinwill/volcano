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

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"
	"k8s.io/utils/cpuset"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/numaaware/policy"
)

type gpuMng struct{}

func NewGPUProvider() policy.HintProvider {
	return &gpuMng{}
}

func (m *gpuMng) Name() string {
	return "gpuMng"
}

// guaranteedGPUs return the intger num of request gpu
func guaranteedGPUs(container *v1.Container) int {
	gpuQuantity := container.Resources.Requests[api.GPUResourceName]
	if gpuQuantity.MilliValue()%1000 != 0 {
		return 0
	}

	return int(gpuQuantity.Value())
}

func (m *gpuMng) GetTopologyHints(container *v1.Container,
	topoInfo *api.NumatopoInfo, resNumaSets api.ResNumaSets) map[string][]policy.TopologyHint {

	// 获取请求的 GPU 数量
	if _, ok := container.Resources.Requests[api.GPUResourceName]; !ok {
		klog.Warningf("container %s has no gpu request", container.Name)
		return nil
	}
	requestNum := guaranteedGPUs(container)
	if requestNum == 0 {
		klog.Warningf(" the gpu request isn't  integer in container %s", container.Name)
		return nil
	}

	availableGPUSet, ok := resNumaSets[api.GPUResourceName]
	if !ok {
		klog.Warningf("no gpu resource")
		return nil
	}
	klog.V(4).Infof("requested: %d, availableGPUSet: %v", requestNum, availableGPUSet)

	return map[string][]policy.TopologyHint{
		api.GPUResourceName: generateGPUTopologyHints(availableGPUSet, topoInfo.GPUDetail, requestNum),
	}
}

// generateCPUTopologyHints return the numa topology hints based on
// - availableGPUs
func generateGPUTopologyHints(availableGPUs cpuset.CPUSet, GPUDetails api.GPUDetails, request int) []policy.TopologyHint {
	minAffinitySize := GPUDetails.NUMANodes().Size()
	hints := []policy.TopologyHint{}
	//Traverse the combination of all NUMA nodes. (For example, [0], [1], and [0, 1]), execute judgment logic for each combination
	bitmask.IterateBitMasks(GPUDetails.NUMANodes().List(), func(mask bitmask.BitMask) {
		// First, update minAffinitySize for the current request size.
		gpusInMask := GPUDetails.GPUsInNUMANodes(mask.GetBits()...).Size()
		if gpusInMask >= request && mask.Count() < minAffinitySize {
			minAffinitySize = mask.Count()
		}

		// Then check to see if we have enough GPUs available on the current
		// numa node bitmask to satisfy the GPU request.
		numMatching := 0
		// Finally, check to see if enough available GPUs remain on the current
		// NUMA node combination to satisfy the GPU request.
		for _, c := range availableGPUs.List() {
			key := fmt.Sprintf("nvidia%d", c)
			if info, ok := GPUDetails[key]; ok {
				if mask.IsSet(info.NUMANodeID) {
					numMatching++
				}
			}
		}

		// If they don't, then move onto the next combination.
		if numMatching < request {
			return
		}

		// Otherwise, create a new hint from the numa node bitmask and add it to the
		// list of hints.  We set all hint preferences to 'false' on the first
		// pass through.
		hints = append(hints, policy.TopologyHint{
			NUMANodeAffinity: mask,
			Preferred:        false,
		})
	})

	// Loop back through all hints and update the 'Preferred' field based on
	// counting the number of bits sets in the affinity mask and comparing it
	// to the minAffinitySize. Only those with an equal number of bits set (and
	// with a minimal set of numa nodes) will be considered preferred.
	for i := range hints {
		if hints[i].NUMANodeAffinity.Count() == minAffinitySize {
			hints[i].Preferred = true
		}
	}

	return hints
}

func (m *gpuMng) Allocate(container *v1.Container,
	bestHit *policy.TopologyHint,
	topoInfo *api.NumatopoInfo,
	resNumaSets api.ResNumaSets) map[string]cpuset.CPUSet {

	requestNum := guaranteedGPUs(container)
	availableGPUSet := resNumaSets[api.GPUResourceName]

	klog.V(4).Infof("alignedCPUs: %v requestNum: %v bestHit %v", availableGPUSet, requestNum, bestHit)

	gputopo := &api.GPUTopology{
		NumGPUs:    topoInfo.GPUDetail.GPUs().Size(),
		GPUDetails: topoInfo.GPUDetail,
	}

	result := cpuset.New()
	if bestHit.NUMANodeAffinity != nil {
		alignedGPUs := cpuset.New()
		for _, numaNodeID := range bestHit.NUMANodeAffinity.GetBits() {
			alignedGPUs = alignedGPUs.Union(availableGPUSet.Intersection(gputopo.GPUDetails.GPUsInNUMANodes(numaNodeID)))
		}

		numAlignedToAlloc := alignedGPUs.Size()
		if requestNum < numAlignedToAlloc {
			numAlignedToAlloc = requestNum
		}

		alignedGPUs, err := takeByTopology(gputopo, alignedGPUs, numAlignedToAlloc)
		if err != nil {
			return map[string]cpuset.CPUSet{
				api.GPUResourceName: cpuset.New(),
			}
		}

		result = result.Union(alignedGPUs)
	}

	// Get any remaining GPUs from what's leftover after attempting to grab aligned ones.
	remainingCPUs, err := takeByTopology(gputopo, availableGPUSet.Difference(result), requestNum-result.Size())
	if err != nil {
		return map[string]cpuset.CPUSet{
			api.GPUResourceName: cpuset.New(),
		}
	}

	result = result.Union(remainingCPUs)

	return map[string]cpuset.CPUSet{
		api.GPUResourceName: result,
	}
}
