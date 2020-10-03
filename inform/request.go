/*
 * Copyright (c) 2017 ZAP Québec
 * Copyright (c) 2020 Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package inform

import (
	"encoding/json"
	"fmt"
)

type SpeedTestStatus struct {
	Latency        uint64  `json:"latency"`
	RunDate        uint64  `json:"rundate"`
	RunTime        uint64  `json:"runtime"`
	StatusDownload uint64  `json:"status_download"`
	StatusPing     uint64  `json:"status_ping"`
	StatusUpload   uint64  `json:"status_upload"`
	XputDownload   float64 `json:"xput_download"`
	XputUpload     float64 `json:"xput_upload"`
}

type EthernetTableEntry struct {
	Name    string `json:"name"`
	Mac     string `json:"mac"`
	NumPort uint64 `json:"num_port"`
}

const NetworkConfigDisabled = "disabled"
const NetworkConfigDhcp = "dhcp"
const NetworkConfigStatic = "static"

type NetworkConfig struct {
	Type    string `json:"type"`
	Ip      string `json:"ip,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	Dns1    string `json:"dns1,omitempty"`
	Dns2    string `json:"dns2,omitempty"`
	IfName  string `json:"ifname,omitempty"`
}

func (n NetworkConfig) MarshalJSON() ([]byte, error) {
	switch n.Type {
	case NetworkConfigDhcp, NetworkConfigDisabled:
		return json.Marshal(&struct {
			Type   string `json:"type,omitempty"`
			IfName string `json:"ifname,omitempty"`
		}{
			Type:   n.Type,
			IfName: n.IfName,
		})
	case NetworkConfigStatic:
		if len(n.Ip) == 0 {
			return nil, fmt.Errorf("invalid IP")
		}
		if len(n.Netmask) == 0 {
			return nil, fmt.Errorf("invalid netmask")
		}
		if len(n.Gateway) == 0 {
			return nil, fmt.Errorf("invalid gateway")
		}
	default:
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
		}{
			Type: NetworkConfigDisabled,
		})

	}
	return json.Marshal(&struct {
		Type    string `json:"type"`
		Ip      string `json:"ip,omitempty"`
		Netmask string `json:"netmask,omitempty"`
		Gateway string `json:"gateway,omitempty"`
		Dns1    string `json:"dns1,omitempty"`
		Dns2    string `json:"dns2,omitempty"`
		IfName  string `json:"ifname,omitempty"`
	}{
		Type:    n.Type,
		Ip:      n.Ip,
		Netmask: n.Netmask,
		Gateway: n.Gateway,
		Dns1:    n.Dns1,
		Dns2:    n.Dns2,
		IfName:  n.IfName,
	})
}

type Interface struct {
	FullDuplex  bool         `json:"full_duplex"`
	Ip          string       `json:"ip"`
	Mac         HardwareAddr `json:"mac"`
	Name        string       `json:"name"`
	Netmask     string       `json:"netmask"`
	NumPort     int          `json:"num_port"`
	RxBytes     uint64       `json:"rx_bytes"`
	RxDropped   uint64       `json:"rx_dropped"`
	RxErrors    uint64       `json:"rx_errors"`
	RxMulticast int          `json:"rx_multicast"`
	RxPackets   uint64       `json:"rx_packets"`
	Speed       uint64       `json:"speed"`
	TxBytes     uint64       `json:"tx_bytes"`
	TxDropped   uint64       `json:"tx_dropped"`
	TxErrors    uint64       `json:"tx_errors"`
	TxPackets   uint64       `json:"tx_packets"`
	Up          bool         `json:"up"`
	Enabled     bool         `json:"enabled"`
	Drops       uint64       `json:"drops"`
	Latency     uint64       `json:"latency"`
	Uptime      uint64       `json:"uptime"`
	Nameservers []string     `json:"namservers"`
	Gateways    []string     `json:"gateways"`
}

type Port struct {
	IfName string `json:"ifname"`
	Name   string `json:"name"`
}

type Inform struct {
	BoardRevision     int           `json:"board_rev,omitempty"`
	BootRomVersion    string        `json:"bootrom_version"`
	ConfigVersion     string        `json:"cfgversion"`
	ConfigNetworkWan  NetworkConfig `json:"config_network_wan,omitempty"`
	ConfigNetworkWan2 NetworkConfig `json:"config_network_wan2,omitempty"`

	//CountryCode          int         `json:"country_code"`
	Default           bool `json:"default"`
	DiscoveryResponse bool `json:"discovery_response"`
	//Fingerprint          string      `json:"fingerprint"`
	EthernetTable           []EthernetTableEntry `json:"ethernet_table,omitempty"`
	FirmwareCapabilities    int32                `json:"fw_caps"`
	GuestToken              string               `json:"guest_token,omitempty"`
	HasDefaultRouteDistance bool                 `json:"has_default_route_distance"`
	HasHostfileUpdate       bool                 `json:"has_dnsmasq_hostfile_update"`
	HasDpi                  bool                 `json:"has_dpi"`
	HasEth1                 bool                 `json:"has_eth1"`
	HasPortA                bool                 `json:"has_porta"`
	HasSshDisable           bool                 `json:"has_ssh_disable"`
	HasVti                  bool                 `json:"has_vti"`
	//HasSpeaker           bool   `json:"has_speaker"` // Not present in rust implementation
	Hostname     string       `json:"hostname"`
	IntfTable    []Interface  `json:"if_table"`
	InformUrl    string       `json:"inform_url"`
	InformIp     string       `json:"inform_ip"`
	Ip           string       `json:"ip"`
	Isolated     bool         `json:"isolated"`
	LastError    string       `json:"last_error,omitempty"` // Not present in rust implementation
	Locating     bool         `json:"locating"`
	Mac          HardwareAddr `json:"mac"`
	Model        string       `json:"model"`
	ModelDisplay string       `json:"model_display"`
	Netmask      string       `json:"netmask"`
	QrId         string       `json:"qrid,omitempty"`
	//RadioTable           []Radio     `json:"radio_table"`
	PortTable          []Port   `json:"config_port_table"`
	RadiusCapabilities int32    `json:"radius_caps"`
	RequiredVersion    string   `json:"required_version"`
	SelfrunBeacon      bool     `json:"selfrun_beacon"`
	Serial             string   `json:"serial"`
	SpectrumScanning   bool     `json:"spectrum_scanning,omitempty"`
	State              int      `json:"state"`
	StreamToken        string   `json:"stream_token,omitempty"`
	SysStats           SysStats `json:"system-stats"`
	Time               int64    `json:"time"`
	Uplink             string   `json:"uplink"`
	Uptime             uint64   `json:"uptime"`
	//VApTable             []VAp       `json:"vap_table"`
	Version string `json:"version"`
	//WifiCapabilities     int         `json:"wifi_caps"`

	// Notify fields
	InformAsNotify bool   `json:"inform_as_notify,omitempty"`
	NotifyReason   string `json:"notif_reason,omitempty"`
	NotifyPayload  string `json:"notif_payload,omitempty"`

	// Speed test
	SpeedtestStatus *SpeedTestStatus `json:"speedtest-status,omitempty"`
}

func (r Inform) Marshal() []byte {
	result, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return nil
	}
	return result
}

func (r Inform) String() string {
	s, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return ""
	}
	return fmt.Sprint(s)
}
