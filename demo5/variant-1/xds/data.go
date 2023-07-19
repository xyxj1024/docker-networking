package main

type Input struct {
	Data []Data `json:"input"`
}

type Data struct {
	UpstreamHost string `json:"upstreamHost"`
	ListenerPort uint32 `json:"listenerPort"`
	ClusterName  string `json:"clusterName"`
}
