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
package dockerutil

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/ocischema"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/uber/kraken/core"
	"github.com/uber/kraken/utils/log"
)

func ParseManifest(r io.Reader) (distribution.Manifest, core.Digest, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, core.Digest{}, fmt.Errorf("read: %s", err)
	}

	manifest, d, err := ParseManifestV2(b)
	if err == nil {
		return manifest, d, err
	}

	// Retry with v2 manifest list.
	return ParseManifestV2List(b)
}

// ParseManifestV2 returns a parsed v2 manifest and its digest.
func ParseManifestV2(bytes []byte) (distribution.Manifest, core.Digest, error) {
	manifest, desc, err := distribution.UnmarshalManifest(schema2.MediaTypeManifest, bytes)
	var ociMediaType bool
	if err != nil {
		ociMediaType = true
		manifest, desc, err = distribution.UnmarshalManifest(v1.MediaTypeImageManifest, bytes)
		if err != nil {
			return nil, core.Digest{}, fmt.Errorf("unmarshal manifest: %s", err)
		} else {
			log.Debugf("unmarshal oci manifest success with mediaType: %s", v1.MediaTypeImageManifest)
		}
	} else {
		log.Debugf("unmarshal docker manifest success with mediaType: %s", schema2.MediaTypeManifest)
	}

	var version int
	var expectedErr error
	if ociMediaType {
		deserializedManifest, ok := manifest.(*ocischema.DeserializedManifest)
		if ok {
			version = deserializedManifest.Manifest.Versioned.SchemaVersion
		} else {
			expectedErr = errors.New("expected ocischema.DeserializedManifest")
		}
	} else {
		deserializedManifest, ok := manifest.(*schema2.DeserializedManifest)
		if ok {
			version = deserializedManifest.Manifest.Versioned.SchemaVersion
		} else {
			expectedErr = errors.New("expected schema2.DeserializedManifest")
		}
	}
	if expectedErr != nil {
		return nil, core.Digest{}, expectedErr
	}

	if version != 2 {
		return nil, core.Digest{}, fmt.Errorf("unsupported manifest version: %d", version)
	}
	d, err := core.ParseSHA256Digest(string(desc.Digest))
	if err != nil {
		return nil, core.Digest{}, fmt.Errorf("parse digest: %s", err)
	}
	return manifest, d, nil
}

// ParseManifestV2List returns a parsed v2 manifest list and its digest.
func ParseManifestV2List(bytes []byte) (distribution.Manifest, core.Digest, error) {
	manifestList, desc, err := distribution.UnmarshalManifest(manifestlist.MediaTypeManifestList, bytes)
	if err != nil {
		manifestList, desc, err = distribution.UnmarshalManifest(v1.MediaTypeImageIndex, bytes)
		if err != nil {
			return nil, core.Digest{}, fmt.Errorf("unmarshal manifestlist: %s", err)
		} else {
			log.Debugf("unmarshal oci index success with mediaType: %s", v1.MediaTypeImageIndex)
		}
	} else {
		log.Debugf("unmarshal docker manifestlist success with mediaType: %s", manifestlist.MediaTypeManifestList)
	}
	deserializedManifestList, ok := manifestList.(*manifestlist.DeserializedManifestList)
	if !ok {
		return nil, core.Digest{}, errors.New("expected manifestlist.DeserializedManifestList")
	}
	version := deserializedManifestList.ManifestList.Versioned.SchemaVersion
	if version != 2 {
		return nil, core.Digest{}, fmt.Errorf("unsupported manifest list version: %d", version)
	}
	d, err := core.ParseSHA256Digest(string(desc.Digest))
	if err != nil {
		return nil, core.Digest{}, fmt.Errorf("parse digest: %s", err)
	}
	return manifestList, d, nil
}

// GetManifestReferences returns a list of references by a V2 manifest
func GetManifestReferences(manifest distribution.Manifest) ([]distribution.Descriptor, error) {
	var refs []distribution.Descriptor
	for _, desc := range manifest.References() {
		_, err := core.ParseSHA256Digest(string(desc.Digest))
		if err != nil {
			return nil, fmt.Errorf("parse digest: %s", err)
		}
		refs = append(refs, distribution.Descriptor{
			Digest:    desc.Digest,
			MediaType: desc.MediaType,
		})
	}
	return refs, nil
}

func GetSupportedManifestTypes() string {
	return fmt.Sprintf("%s,%s,%s,%s", schema2.MediaTypeManifest, manifestlist.MediaTypeManifestList, v1.MediaTypeImageManifest, v1.MediaTypeImageIndex)
}
