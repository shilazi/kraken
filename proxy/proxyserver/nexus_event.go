package proxyserver

import "time"

//{
//    "action": "CREATED",
//    "asset": {
//        "name": "v2/test/image/manifests/sha256:1093ee5c7c0ecb8e6ec21645fb5ad4a94da67d70ae1d4ac6977423f8f275fb96",
//        "id": "a4979ccf72dc712842cbc453e9322566",
//        "assetId": "ZG9ja2VyLXNuYXBzaG90czphNDk3OWNjZjcyZGM3MTI4NDJjYmM0NTNlOTMyMjU2Ng",
//        "format": "docker"
//    },
//    "timestamp": "2024-01-19T02:08:27.461+00:00",
//    "nodeId": "5D9EDA16-682627BC-40903AE8-831403D3-08150DA0",
//    "initiator": "admin/192.168.0.2",
//    "repositoryName": "docker-testrepo"
//}

// TODO: Component event not contain 'digest', use asset event with filter aim to preheat.
//		Condition: action == 'CREATED' && asset.format == 'docker' && asset.name indexOf '/v2/-/blobs/' == -1

// NexusEvent holds action about the target of an event.
type NexusEvent struct {
	Timestamp      time.Time
	NodeId         string
	Initiator      string
	RepositoryName string
	Action         string
	Asset          *NexusAsset
}

// NexusAsset holds information about the target of an asset.
type NexusAsset struct {
	Id      string
	Format  string
	Name    string
	AssetId string
}
