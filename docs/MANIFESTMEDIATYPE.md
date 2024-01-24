**Table of Contents**

- [Docker manifest.v2+json](#docker-manifestv2json)
    - [Legacy docker build](#legacy-docker-build)
    - [Buildx version < 0.10.0](#buildx-version-0100)
    - [Buildx version >= 0.10.0](#buildx-version-0100_1)
- [Docker manifest.list.v2+json](#docker-manifestlistv2json)
    - [Buildx version < 0.10.0](#buildx-version-0100_2)
    - [Buildx version >= 0.10.0](#buildx-version-0100_3)
- [OCI index.v1+json](#oci-indexv1json)
    - [Buildx version >= 0.10.0](#buildx-version-0100_4)
- [Summary](#summary)

# Manifest Media Type

With the emergence of the media type defined by [Open Container Initiative (OCI) Specifications](https://github.com/opencontainers/image-spec), [Docker Image Manifest V 2, Schema 2](https://distribution.github.io/distribution/spec/manifest-v2-2/) defined media type has become a thing of the past. 

Modern container image build tools now use the new [OCI Image Manifest Specification](https://github.com/opencontainers/image-spec/blob/main/manifest.md) to generate manifest content.

**Import:!!** It is worth noting that when using [buildx](https://github.com/docker/buildx) versions below [0.10.0](https://docs.docker.com/build/attestations/), [Docker Image Manifest V 2, Schema 2](https://distribution.github.io/distribution/spec/manifest-v2-2/)  specifications are followed. Starting from version [0.10.0](https://github.com/docker/buildx/releases/tag/v0.10.0) onwards, the default specification followed is [Open Container Initiative (OCI) Specifications](https://github.com/opencontainers/image-spec).

## Docker manifest.v2+json

These build will make  `application/vnd.docker.distribution.manifest.v2+json` media type.

### Legacy Docker Build

```bash
docker build --push -t docker.io/fakeone/fakeimage:v0.0.1 .
```

### [Buildx](https://github.com/docker/buildx/issues/1513) Version < 0.10.0

Docker buildx build without `--platform` option:

```bash
docker buildx build --push -t docker.io/fakeone/fakeimage:v0.0.1 .
```

### [Buildx](https://github.com/docker/buildx/issues/1513) Version >= 0.10.0

Docker buildx build without `--platform` and  with `--sbom=false --provenance=false` options:

```bash
docker buildx build --push -t docker.io/fakeone/fakeimage:v0.0.1 . --sbom=false --provenance=false
```

Example showing an image manifest:

```json
{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 600,
      "digest": "sha256:26fc80ffe8cc20647e054e923"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 3408480,
         "digest": "sha256:26fc80ffe8cc20647e054e923"
      }
   ]
}
```

## Docker manifest.list.v2+json

These build will make  `application/vnd.docker.distribution.manifest.list.v2+json` media type.

### [Buildx](https://github.com/docker/buildx/issues/1513) Version < 0.10.0

Docker buildx build with `--platform` option:

```bash
docker buildx build --push -t docker.io/fakeone/fakeimage:v0.0.1 . --platform linux/amd64,linux/arm64
```

### [Buildx](https://github.com/docker/buildx/issues/1513) Version >= 0.10.0

Docker buildx build with `--platform` and  with `--sbom=false --provenance=false` options:

```bash
docker buildx build --push -t docker.io/fakeone/fakeimage:v0.0.1 . --platform linux/amd64,linux/arm64 \
    --sbom=false --provenance=false
```

Example showing an image manifest:

```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "manifests": [
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:26fc80ffe8cc20647e054e923",
      "size": 502,
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:26fc80ffe8cc20647e054e923",
      "size": 502,
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      }
    }
  ]
}
```

## OCI index.v1+json

These build will make  `application/vnd.oci.image.index.v1+json` media type.

### [Buildx](https://github.com/docker/buildx/issues/1513) Version >= 0.10.0

Docker buildx build without `--platform` and  without `--sbom=false --provenance=false` options:

```bash
docker buildx build --push -t docker.io/fakeone/fakeimage:v0.0.1 .
```

Docker buildx build with `--platform` and  without `--sbom=false --provenance=false` option:

```bash
docker buildx build --push -t docker.io/fakeone/fakeimage:v0.0.1 . --platform linux/amd64,linux/arm64
```

Example showing two image manifest:

> without `--platform` option

```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:26fc80ffe8cc20647e054e923",
      "size": 480,
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:26fc80ffe8cc20647e054e923",
      "size": 566,
      "annotations": {
        "vnd.docker.reference.digest": "sha256:26fc80ffe8cc20647e054e923",
        "vnd.docker.reference.type": "attestation-manifest"
      },
      "platform": {
        "architecture": "unknown",
        "os": "unknown"
      }
    }
  ]
}
```

>  with `--platform` option

```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:26fc80ffe8cc20647e054e923",
      "size": 480,
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:329e450e74602cc8eff08cf62",
      "size": 480,
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:126fc80ffe8cc20647e054e92",
      "size": 566,
      "annotations": {
        "vnd.docker.reference.digest": "sha256:26fc80ffe8cc20647e054e923",
        "vnd.docker.reference.type": "attestation-manifest"
      },
      "platform": {
        "architecture": "unknown",
        "os": "unknown"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:1329e450e74602cc8eff08cf6",
      "size": 566,
      "annotations": {
        "vnd.docker.reference.digest": "sha256:329e450e74602cc8eff08cf62",
        "vnd.docker.reference.type": "attestation-manifest"
      },
      "platform": {
        "architecture": "unknown",
        "os": "unknown"
      }
    }
  ]
}
```

## Summary

The above is the current situation based on testing, and it is anticipated that all images in the future will adhere to the [OCI Specifications](https://github.com/opencontainers/image-spec). This is something to look forward to.
