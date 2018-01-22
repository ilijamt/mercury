package config

import (
	"github.com/schubergphilis/mercury/src/balancer"
	"github.com/schubergphilis/mercury/src/proxy"
)

// ProxyBackendNodeUpdate contains backend updates to proxy
type ProxyBackendNodeUpdate struct {
	//ClusterNode string            `json:"clusternode"`
	PoolName        string            `json:"poolname"`
	BackendName     string            `json:"backendname"`
	BackendNodeUUID string            `toml:"uuid" json:"uuid"`
	BackendNode     proxy.BackendNode `toml:"backendnode" json:"backendnode"`
}

// ProxyBackendStatisticsUpdate contains backend updates to proxy
type ProxyBackendStatisticsUpdate struct {
	//ClusterNode string            `json:"clusternode"`
	ClusterNode string              `json:"clusternode"`
	DNSEntry    DNSEntry            `json:"dnsentry"`
	PoolName    string              `json:"poolname"`
	BackendName string              `json:"backendname"`
	Statistics  balancer.Statistics `toml:"statistics" json:"statistics"`
}

// ClusterPacketGlobalDNSUpdate contains a dns update with records
type ClusterPacketGlobalDNSUpdate struct {
	ClusterNode string      `json:"clusternode"`
	PoolName    string      `json:"poolname"`
	BackendName string      `json:"backendname"`
	DNSEntry    DNSEntry    `json:"dnsentry"`
	BalanceMode BalanceMode `json:"balancemode"`
	BackendUUID string      `toml:"uuid" json:"uuid"`
	Online      bool        `toml:"online" json:"online"`
}

// ClusterPacketGlbalDNSStatisticsUpdate contains backend updates to proxy
type ClusterPacketGlbalDNSStatisticsUpdate struct {
	ClusterNode       string    `json:"clusternode"`
	PoolName          string    `json:"poolname"`
	BackendName       string    `json:"backendname"`
	DNSEntry          DNSEntry  `json:"dnsentry"`
	UUID              string    `json:"uuid"`
	ClientsConnected  int64     `json:"clientsconnected"`
	ClientsConnects   int64     `json:"clientsconnects"`
	RX                int64     `json:"rx"`
	TX                int64     `json:"tx"`
	ResponseTimeValue []float64 `json:"responsetimevalue"`
	///Statistics  balancer.Statistics `toml:"statistics" json:"statistics"`
}

// ClusterPacketClearProxyStatistics contains a request to clear the proxy statistics
type ClusterPacketClearProxyStatistics struct {
	PoolName    string `json:"poolname"`
	BackendName string `json:"backendname"`
}

// PacketConfigRequest is the packet type sent for configuration requests
type ClusterPacketConfigRequest struct{}
