/*
Copyright 2021 The Volcano Authors.

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

package policy

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"
)

func Test_restricted_predicate(t *testing.T) {
	testCases := []struct {
		name           string
		providersHints []map[string][]TopologyHint
		expect         TopologyHint
	}{
		{
			name: "test-1",
			providersHints: []map[string][]TopologyHint{
				{
					"cpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
				{
					"gpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
			},
			expect: TopologyHint{
				NUMANodeAffinity: func() bitmask.BitMask {
					mask, _ := bitmask.NewBitMask(0)
					return mask
				}(),
				Preferred: true,
			},
		},
		{
			name: "test-2",
			providersHints: []map[string][]TopologyHint{
				{
					"cpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(1)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
				{
					"gpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(1)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
			},
			expect: TopologyHint{
				NUMANodeAffinity: func() bitmask.BitMask {
					mask, _ := bitmask.NewBitMask(1)
					return mask
				}(),
				Preferred: true,
			},
		},
		{
			name: "test-3",
			providersHints: []map[string][]TopologyHint{
				{
					"cpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0)
								return mask
							}(),
							Preferred: true,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(1)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
				{
					"gpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0)
								return mask
							}(),
							Preferred: true,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(1)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
			},
			expect: TopologyHint{
				NUMANodeAffinity: func() bitmask.BitMask {
					mask, _ := bitmask.NewBitMask(0)
					return mask
				}(),
				Preferred: true,
			},
		},
		{
			name: "test-4",
			providersHints: []map[string][]TopologyHint{
				{
					"cpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
				{
					"gpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(1)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
			},
			expect: TopologyHint{
				NUMANodeAffinity: func() bitmask.BitMask {
					mask, _ := bitmask.NewBitMask(0)
					return mask
				}(),
				Preferred: false,
			},
		},
		{
			name: "test-5",
			providersHints: []map[string][]TopologyHint{
				{
					"cpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: false,
						},
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
				{
					"gpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
			},
			expect: TopologyHint{
				NUMANodeAffinity: func() bitmask.BitMask {
					mask, _ := bitmask.NewBitMask(0)
					return mask
				}(),
				Preferred: true,
			},
		},
		{
			name: "test-6",
			providersHints: []map[string][]TopologyHint{
				{
					"cpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
				{
					"gpu": []TopologyHint{
						{
							NUMANodeAffinity: func() bitmask.BitMask {
								mask, _ := bitmask.NewBitMask(0, 1)
								return mask
							}(),
							Preferred: true,
						},
					},
				},
			},
			expect: TopologyHint{
				NUMANodeAffinity: func() bitmask.BitMask {
					mask, _ := bitmask.NewBitMask(0, 1)
					return mask
				}(),
				Preferred: true,
			},
		},
	}

	for _, testcase := range testCases {
		policy := NewPolicyRestricted([]int{0, 1})
		bestHit, _ := policy.Predicate(testcase.providersHints)
		if !equality.Semantic.DeepEqual(bestHit, testcase.expect) {
			t.Errorf("%s failed, expect %v, bestHit= %v\n", testcase.name, testcase.expect, bestHit)
		}
	}
}
