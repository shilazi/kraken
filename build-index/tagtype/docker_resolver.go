// Copyright (c) 2016-2019 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package tagtype

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/docker/distribution"
	"github.com/uber/kraken/core"
	"github.com/uber/kraken/origin/blobclient"
	"github.com/uber/kraken/utils/dockerutil"
	"github.com/uber/kraken/utils/log"
)

type dockerResolver struct {
	originClient blobclient.ClusterClient
}

// Resolve returns all layers + manifest of given tag as its dependencies.
func (r *dockerResolver) Resolve(tag string, digest core.Digest) (core.DigestList, error) {
	log.Infof("Resolve tag: %s, digest: %+v", tag, digest)
	deps, err := r.recursionResolve("root", tag, digest)
	if err != nil {
		return nil, fmt.Errorf("get manifest references: %s", err)
	}
	log.Debugf("Resolve deps: %+v", deps)
	return append(deps, digest), nil
}

// recursionResolve compatibility docker manifestlist or oci index media type
func (r *dockerResolver) recursionResolve(layer, tag string, digest core.Digest) (core.DigestList, error) {
	log.Infof("recursion resolve layer: %s, tag: %s, digest: %s", layer, tag, digest)
	m, err := r.downloadManifest(tag, digest)
	if err != nil {
		log.Errorf("recursion resolve layer: %s, downloadManifest: %s", layer, err)
		return nil, fmt.Errorf("download manifest: %s", err)
	}
	refs, err := dockerutil.GetManifestReferences(m)
	if err != nil {
		log.Errorf("recursion resolve layer: %s, GetManifestReferences: %s", layer, err)
		return nil, fmt.Errorf("get manifest references: %s", err)
	}
	var deps core.DigestList
	recursionMediaTypes := dockerutil.GetSupportedManifestTypes()
	for i, ref := range refs {
		mediaType := ref.MediaType
		rLayer := fmt.Sprintf("%s-%d", layer, i)
		d, _ := core.ParseSHA256Digest(string(ref.Digest))
		if strings.Contains(recursionMediaTypes, mediaType) {
			recRefs, recErr := r.recursionResolve(rLayer, tag, d)
			if recErr != nil {
				log.Errorf("recursion resolve layer: %s, mediaType: %s, err: %s", rLayer, mediaType, recErr)
				continue
			}
			deps = append(deps, recRefs...)
		} else {
			log.Infof("recursion resolve layer: %s, append digest: %+v", rLayer, ref.Digest)
			deps = append(deps, digest)
		}
	}
	return deps, nil
}

func (r *dockerResolver) downloadManifest(tag string, d core.Digest) (distribution.Manifest, error) {
	buf := &bytes.Buffer{}
	if err := r.originClient.DownloadBlob(tag, d, buf); err != nil {
		return nil, fmt.Errorf("download blob: %s", err)
	}
	manifest, _, err := dockerutil.ParseManifest(buf)
	if err != nil {
		return nil, fmt.Errorf("parse manifest: %s", err)
	}
	return manifest, nil
}
