package proxyserver

import "time"

//{
//    "Events": [
//        {
//            "Id": "YSnFhJWCpNoVynYgHQuyz2xdoyXMI7BD",
//            "TimeStamp": "2024-01-18T14:37:42.194082639Z",
//            "Action": "push",
//            "Target": {
//                "MediaType": "application/vnd.docker.distribution.manifest.v2+json",
//                "Digest": "sha256:3eacb2401d9859bb89af468200241bbd46645c3d63e85f16393759ca4aa1c18e",
//                "Repository": "test/image",
//                "Url": "https://foo.example.com/v2/test/image/manifests/v1.0.2",
//                "Tag": "v1.0.2"
//            },
//            "Request": null,
//            "Actor": null
//        }
//    ]
//}

// TODO: There is a bug about registry event.
// 		Target.MediaType always 'application/vnd.docker.distribution.manifest.v2+json'.
// 		Whatever push with any other media type, such as:
// 		* application/vnd.oci.image.index.v1+json
// 		* application/vnd.docker.distribution.manifest.list.v2+json

// Notification holds all events. refer to https://docs.docker.com/registry/notifications/.
type Notification struct {
	Events []Event
}

// Event holds the details of a event.
type Event struct {
	ID        string `json:"Id"`
	TimeStamp time.Time
	Action    string
	Target    *Target
}

// Target holds information about the target of a event.
type Target struct {
	MediaType  string
	Digest     string
	Repository string
	URL        string `json:"Url"`
	Tag        string
}
