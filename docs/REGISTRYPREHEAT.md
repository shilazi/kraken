**Table of Contents**

- [Registry Notification](#registry-notification)
- [Nexus Repository Webhook](#nexus-repository-webhook)
- [More Event Backends](#more-event-backends)
- [Implementation Notes](#implementation-notes)

# Registry Preheat

Now can preheat image blobs with [registry notifications](https://distribution.github.io/distribution/about/notifications/) and [nexus repository webhook](https://help.sonatype.com/en/enabling-a-repository-webhook-capability.html).

## Registry Notification

When using [Harbor](https://goharbor.io/) or a headless [registry](https://distribution.github.io/distribution/) as the registry, you can configure notifications for the registry. 

Example showing a section registry notification configuration:

```
notifications:
  endpoints:
  - name: kraken
    disabled: false
    url: http://kraken-proxy.p2p.svc.cluster.local:81/registry/notifications
    timeout: 3000ms
    threshold: 5
    backoff: 1s
```

After successfully pushing an image, it will send events to the configured endpoint. 

Example showing a registry notification event:

```json
{
    "Events": [
        {
            "Id": "YSnFhJWCpNoVynYgHQuyz2xdoyXMI7BD",
            "TimeStamp": "2024-01-18T14:37:42.194082639Z",
            "Action": "push",
            "Target": {
                "MediaType": "application/vnd.docker.distribution.manifest.v2+json",
                "Digest": "sha256:26fc80ffe8cc20647e054e923",
                "Repository": "fakeone/fakeimage",
                "Url": "https://foo.example.com/v2/fakeone/fakeimage/manifests/v0.0.1",
                "Tag": "v0.0.1"
            },
            "Request": null,
            "Actor": null
        }
    ]
}
```

After receiving an event, [kraken-proxy will attempt deserialization](https://github.com/shilazi/kraken/blob/master/proxy/proxyserver/preheat.go#L172). If the event meets the criteria of being a **push** event and match the specified **Target.MediaType** regExp, it will enter the preheating [processing phase](https://github.com/shilazi/kraken/blob/master/proxy/proxyserver/preheat.go#L94).

NOTE: It seems to be an area for optimization since the media type is currently always `application/vnd.docker.distribution.manifest.v2+json`.

## Nexus Repository Webhook

When using [Sonatype Nexus OSS](https://www.sonatype.com/products/sonatype-nexus-oss)  as the registry, you can configure webhook for the repository.

The path of the repository webhook management interface configuration:

* System -> Capabilities -> Create capabilities -> Select Capability Type -> Webhook: Repository
    * Repository: select **docker format** repository，such as `docker-testrepo`
    * Event Types: **asset**
    * URL: such as `http://kraken-proxy.p2p.svc.cluster.local:81/registry/notifications`
    * Secret Key:  write casually, as long as you like, such as `test`

After successfully pushing an image, it will send events to the configured webhook URL. 

Example showing a nexus repository webhook event:

```json
{
    "action": "CREATED",
    "asset": {
        "name": "v2/fakeone/fakeimage/manifests/sha256:26fc80ffe8cc20647e054e923",
        "id": "a4979ccf72dc7128",
        "assetId": "ZG9ja2VyLXNuYXBzaG90czphN",
        "format": "docker"
    },
    "timestamp": "2024-01-18T14:37:42+00:00",
    "nodeId": "5D9EDA16-682627BC-40903AE8",
    "initiator": "awesom/192.168.0.2",
    "repositoryName": "docker-testrepo"
}
```

After receiving an event, [kraken-proxy will attempt deserialization](https://github.com/shilazi/kraken/blob/master/proxy/proxyserver/preheat.go#L211). If the event meets the criteria of being a **CREATED** event and match the specified **asset.name** regExp, it will enter the preheating [processing phase](https://github.com/shilazi/kraken/blob/master/proxy/proxyserver/preheat.go#L94).

NOTE: Many events are sent, and there are also duplicate SHA256 values, but fortunately, preheating operations can proceed normally.

## More Event Backends

If you are not using either of the above two options, you can explore the [source code](https://github.com/shilazi/kraken/blob/master/proxy/proxyserver/preheat.go) and tailor the development to your specific needs.

## Implementation Notes

When encountering a new [Open Container Initiative (OCI) Specifications](https://github.com/opencontainers/image-spec) manifest, it always starts with the [application/vnd.oci.image.index.v1+json](https://github.com/opencontainers/image-spec/blob/main/image-index.md) format. Therefore, recursive processing is required until the [blob layer](https://github.com/shilazi/kraken/blob/master/proxy/proxyserver/preheat.go#L122) are parsed.
