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
package proxyserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof" // Registers /debug/pprof endpoints in http.DefaultServeMux.
	"regexp"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/uber/kraken/core"
	"github.com/uber/kraken/origin/blobclient"
	"github.com/uber/kraken/utils/dockerutil"
	"github.com/uber/kraken/utils/handler"
	"github.com/uber/kraken/utils/httputil"
	"github.com/uber/kraken/utils/log"
)

var (
	_manifestApiRegexp       = regexp.MustCompile(`^v2/.*/sha256:[a-z0-9]+`)
	_nexusBlobApiRegexp      = regexp.MustCompile(`^v2/-/blobs/sha256:[a-z0-9]+`)
	_imageMediaTypeRegexp    = regexp.MustCompile(`^application/vnd\.(oci|docker)\..*`)
	_ociImageLayerRegexp     = regexp.MustCompile(`^application/vnd\.oci\.image\.layer\..*`)
	_dockerImageRootfsRegexp = regexp.MustCompile(`^application/vnd\.docker\.image\.rootfs\..*`)
)

// PreheatHandler defines the handler of preheat.
type PreheatHandler struct {
	clusterClient blobclient.ClusterClient
}

// preheatImage defines the repo and digest of preheat image.
type preheatImage struct {
	Repository string
	Digest     string
}

// NewPreheatHandler creates a new preheat handler.
func NewPreheatHandler(client blobclient.ClusterClient) *PreheatHandler {
	return &PreheatHandler{client}
}

// Handle notifies origins to cache the blob related to the image.
func (ph *PreheatHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	// use map receive data first
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return handler.Errorf("decode body: %s", err)
	}

	// then marshal data to byte again, use different event backend parse
	b, _ := json.Marshal(data)
	preheatImages, err := parseRegistryEvent(b)
	if err != nil {
		preheatImages, err = parseNexusEvent(b)
		if err != nil {
			return handler.Errorf("parse event: %s", err)
		} else {
			log.Debugf("parse nexus event success")
		}
	} else {
		log.Debugf("parse registry event success")
	}

	for _, image := range preheatImages {
		repo := image.Repository
		digest := image.Digest

		log.With("repo", repo, "digest", digest).Infof("deal push image event")
		err := ph.process(repo, digest)
		if err != nil {
			log.With("repo", repo, "digest", digest).Errorf("handle preheat: %s", err)
		}
	}
	return nil
}

func (ph *PreheatHandler) process(repo, digest string) error {
	return ph.recursionProcess("root", repo, digest)
}

func (ph *PreheatHandler) recursionProcess(layer, repo, digest string) error {
	log.Infof("recursion process layer %s, repo: %s, digest: %s", layer, repo, digest)
	manifest, err := ph.fetchManifest(repo, digest)
	if err != nil {
		log.Errorf("recursion process layer %s, fetchManifest: %s", layer, err)
		return fmt.Errorf("fetch manifest: %s", err)
	}
	refs, err := dockerutil.GetManifestReferences(manifest)
	if err != nil {
		log.Errorf("recursion process layer %s, GetManifestReferences: %s", layer, err)
		return fmt.Errorf("get manifest references: %s", err)
	}
	recursionMediaTypes := dockerutil.GetSupportedManifestTypes()
	for i, ref := range refs {
		mediaType := ref.MediaType
		rLayer := fmt.Sprintf("%s-%d", layer, i)
		rLayerInfo := fmt.Sprintf("recursion process layer: %s, mediaType: %s", rLayer, mediaType)
		d, _ := core.ParseSHA256Digest(string(ref.Digest))
		switch {
		case strings.Contains(recursionMediaTypes, mediaType):
			err = ph.recursionProcess(rLayer, repo, d.String())
			if err != nil {
				log.Errorf("%s, err: %s", rLayerInfo, err)
			}
		case _ociImageLayerRegexp.MatchString(mediaType) || _dockerImageRootfsRegexp.MatchString(mediaType):
			go func(repo string, digest core.Digest) {
				log.With("repo", repo, "digest", digest).Infof("trigger origin cache: %+v", digest)
				_, err = ph.clusterClient.GetMetaInfo(repo, digest)
				if err != nil && !httputil.IsAccepted(err) {
					log.With("repo", repo, "digest", digest).Errorf("notify origin cache: %s", err)
				}
			}(repo, d)
		default:
			log.Debugf("%s, not image layer, omit", rLayerInfo, err)
		}
	}
	return nil
}

func (ph *PreheatHandler) fetchManifest(repo, digest string) (distribution.Manifest, error) {
	d, err := core.ParseSHA256Digest(digest)
	if err != nil {
		return nil, fmt.Errorf("Error parse digest: %s ", err)
	}

	buf := &bytes.Buffer{}
	// there may be a gap between registry finish uploading manifest and send notification.
	// see https://github.com/docker/distribution/issues/2625.
	interval := 100 * time.Millisecond
	for i := 0; i < 4; i++ {
		if i != 0 {
			time.Sleep(interval)
			interval = interval * 2
		}
		if err := ph.clusterClient.DownloadBlob(repo, d, buf); err == nil {
			break
		} else if err == blobclient.ErrBlobNotFound {
			continue
		} else {
			return nil, fmt.Errorf("download manifest: %s", err)
		}
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("manifest not found")
	}

	manifest, _, err := dockerutil.ParseManifest(buf)
	if err != nil {
		return nil, fmt.Errorf("parse manifest: %s", err)
	}
	return manifest, nil
}

// parseRegistryEvent parse to registry struct with event data
func parseRegistryEvent(b []byte) ([]*preheatImage, error) {
	var notification Notification
	if err := json.Unmarshal(b, &notification); err != nil {
		return nil, fmt.Errorf("unmarshal registry event: %s", err)
	}

	events := filterRegistryEvents(&notification)
	if len(events) == 0 {
		return nil, fmt.Errorf("registry event valid list is empty")
	}

	var preheatImages []*preheatImage
	for _, event := range events {
		preheatImages = append(preheatImages, &preheatImage{
			Repository: event.Target.Repository,
			Digest:     event.Target.Digest,
		})
	}
	return preheatImages, nil
}

// filterRegistryEvents pick out the push manifest events.
func filterRegistryEvents(notification *Notification) []*Event {
	var events []*Event
	for _, event := range notification.Events {
		isManifest := _imageMediaTypeRegexp.MatchString(event.Target.MediaType)
		if !isManifest {
			continue
		}

		if event.Action == "push" {
			events = append(events, &event)
			continue
		}
	}
	return events
}

//parseNexusEvent parse to nexus struct with event data
func parseNexusEvent(b []byte) ([]*preheatImage, error) {
	var nexusEvent NexusEvent
	if err := json.Unmarshal(b, &nexusEvent); err != nil {
		return nil, fmt.Errorf("unmarshal nexus event: %s", err)
	}

	assetName := filterNexusEvent(&nexusEvent)
	if len(assetName) == 0 {
		return nil, fmt.Errorf("nexus event format or action not match")
	}

	assetSplit := strings.Split(assetName, "/manifests/")
	if _, err := core.ParseSHA256Digest(assetSplit[1]); err != nil {
		return nil, fmt.Errorf("nexus event parse digest: %s", err)
	}

	preheatImages := []*preheatImage{
		{
			Repository: assetSplit[0],
			Digest:     assetSplit[1],
		},
	}
	return preheatImages, nil
}

// filterNexusEvent pick out the CREATED docker format event.
func filterNexusEvent(nexusEvent *NexusEvent) string {
	validAction := nexusEvent.Action == "CREATED"
	validFormat := nexusEvent.Asset != nil && nexusEvent.Asset.Format == "docker"

	if !validAction || !validFormat {
		return ""
	}

	assetName := nexusEvent.Asset.Name
	if _nexusBlobApiRegexp.MatchString(assetName) {
		return ""
	}

	if !_manifestApiRegexp.MatchString(assetName) {
		return ""
	}
	return strings.Replace(assetName, "v2/", "", 1)
}
