/*
Copyright 2022 The OpenYurt Authors.

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

package disk

import (
	"os"
	"testing"

	"github.com/openyurtio/openyurt/pkg/yurthub/storage"
)

var keyFuncTestDir = "/tmp/oy-diskstore-keyfunc"

func TestKeyFunc(t *testing.T) {
	cases := map[string]struct {
		info   storage.KeyBuildInfo
		key    string
		err    error
		isRoot bool
	}{
		"namespaced resource key": {
			info: storage.KeyBuildInfo{
				Component: "kubelet",
				Resources: "pods",
				Group:     "",
				Version:   "v1",
				Namespace: "kube-system",
				Name:      "kube-proxy-xx",
			},
			key:    "kubelet/pods.v1.core/kube-system/kube-proxy-xx",
			isRoot: false,
		},
		"non-namespaced resource key": {
			info: storage.KeyBuildInfo{
				Component: "kubelet",
				Resources: "nodes",
				Group:     "",
				Version:   "v1",
				Name:      "edge-worker",
			},
			key:    "kubelet/nodes.v1.core/edge-worker",
			isRoot: false,
		},
		"resource list key": {
			info: storage.KeyBuildInfo{
				Component: "kubelet",
				Resources: "pods",
				Group:     "",
				Version:   "v1",
			},
			key:    "kubelet/pods.v1.core",
			isRoot: true,
		},
		"resource list namespace key": {
			info: storage.KeyBuildInfo{
				Component: "kube-proxy",
				Resources: "services",
				Group:     "",
				Version:   "v1",
				Namespace: "default",
			},
			key:    "kube-proxy/services.v1.core/default",
			isRoot: true,
		},
		"get resources in apps group": {
			info: storage.KeyBuildInfo{
				Component: "controller",
				Resources: "deployments",
				Group:     "apps",
				Version:   "v1",
				Namespace: "default",
				Name:      "nginx",
			},
			key:    "controller/deployments.v1.apps/default/nginx",
			isRoot: false,
		},
		"get crd resources": {
			info: storage.KeyBuildInfo{
				Component: "controller",
				Resources: "foos",
				Group:     "bars.extension.io",
				Version:   "v1alpha1",
				Namespace: "kube-system",
				Name:      "foobar",
			},
			key:    "controller/foos.v1alpha1.bars.extension.io/kube-system/foobar",
			isRoot: false,
		},
		"no component err key": {
			info: storage.KeyBuildInfo{
				Resources: "nodes",
			},
			err: storage.ErrEmptyComponent,
		},
		"no resource err key": {
			info: storage.KeyBuildInfo{
				Component: "kubelet",
				Name:      "edge-worker",
			},
			err: storage.ErrEmptyResource,
		},
		"get namespace": {
			info: storage.KeyBuildInfo{
				Component: "kubelet",
				Resources: "namespaces",
				Group:     "",
				Version:   "v1",
				Namespace: "kube-system",
				Name:      "kube-system",
			},
			key: "kubelet/namespaces.v1.core/kube-system",
		},
		"list namespace": {
			info: storage.KeyBuildInfo{
				Component: "kubelet",
				Resources: "namespaces",
				Group:     "",
				Version:   "v1",
				Namespace: "",
				Name:      "kube-system",
			},
			key: "kubelet/namespaces.v1.core/kube-system",
		},
	}

	disk, err := NewDiskStorage(keyFuncTestDir)
	if err != nil {
		t.Errorf("failed to create disk store, %v", err)
		return
	}
	keyFunc := disk.KeyFunc
	for c, s := range cases {
		t.Run(c, func(t *testing.T) {
			key, err := keyFunc(s.info)
			if err != s.err {
				t.Errorf("unexpected err for case: %s, want: %s, got: %s", c, err, s.err)
			}

			if err == nil {
				storageKey := key.(storageKey)
				if storageKey.Key() != s.key {
					t.Errorf("unexpected key for case: %s, want: %s, got: %s", c, s.key, storageKey.Key())
				}

				if storageKey.isRootKey() != s.isRoot {
					t.Errorf("unexpected key type for case: %s, want: %v, got: %v", c, s.isRoot, storageKey.isRootKey())
				}
			}
		})
	}
	os.RemoveAll(keyFuncTestDir)
}
