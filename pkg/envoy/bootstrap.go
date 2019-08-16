package envoy

type Bootstrap struct {
	DynamicResources DynamicResources `json:"dynamic_resources"`
	Node             Node             `json:"node"`
	StaticResources  StaticResources  `json:"static_resources"`
	Admin            Admin            `json:"admin"`
}

type Grpc struct {
	ClusterName string `json:"cluster_name"`
}
type GrpcServices struct {
	Grpc Grpc `json:"envoy_grpc"`
}
type AdsConfig struct {
	APIType      string       `json:"api_type"`
	GrpcServices GrpcServices `json:"grpc_services"`
}
type APIConfigSource struct {
	APIType      string       `json:"api_type"`
	GrpcServices GrpcServices `json:"grpc_services"`
}
type CdsConfig struct {
	APIConfigSource APIConfigSource `json:"api_config_source"`
}
type LdsConfig struct {
	APIConfigSource APIConfigSource `json:"api_config_source"`
}
type DynamicResources struct {
	AdsConfig AdsConfig `json:"ads_config"`
	CdsConfig CdsConfig `json:"cds_config"`
	LdsConfig LdsConfig `json:"lds_config"`
}
type Node struct {
	Cluster string `json:"cluster"`
	ID      string `json:"id"`
}
type SocketAddress struct {
	Address   string `json:"address"`
	PortValue int    `json:"port_value"`
}
type Hosts struct {
	SocketAddress SocketAddress `json:"socket_address"`
}
type HTTP2ProtocolOptions struct {
}
type Clusters struct {
	Name                 string               `json:"name"`
	Type                 string               `json:"type"`
	ConnectTimeout       string               `json:"connect_timeout"`
	Hosts                []Hosts              `json:"hosts"`
	HTTP2ProtocolOptions HTTP2ProtocolOptions `json:"http2_protocol_options"`
}
type StaticResources struct {
	Clusters []Clusters `json:"clusters"`
}
type Address struct {
	SocketAddress SocketAddress `json:"socket_address"`
}
type Admin struct {
	AccessLogPath string  `json:"access_log_path"`
	ProfilePath   string  `json:"profile_path"`
	Address       Address `json:"address"`
}
