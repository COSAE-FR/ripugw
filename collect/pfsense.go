/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package collect

import (
	"github.com/COSAE-FR/ripugw/inform"
	"github.com/COSAE-FR/ripugw/pfconf"
	hoststats "github.com/shirou/gopsutil/host"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

type PfSenseTranslationTable struct {
	Wan  string `json:"wan"`
	Wan2 string `json:"wan2,omitempty"`
	Lan  string `json:"lan"`
	Uid  string `json:"uid,omitempty"`
}

type TranslatedInterface struct {
	Physical   inform.Interface
	Pfsense    pfconf.Interface
	UnifiName  string
	UnifiLabel string
}

type PfSenseTranslation struct {
	Wan  TranslatedInterface
	Wan2 TranslatedInterface
	Lan  TranslatedInterface
	Uid  TranslatedInterface
}

func prepareRealInterface(pfInterface pfconf.Interface, interfaces []inform.Interface, unifiName string, unifiLabel string) TranslatedInterface {
	result := TranslatedInterface{
		Pfsense:    pfInterface,
		UnifiName:  unifiName,
		UnifiLabel: unifiLabel,
	}
	for _, iface := range interfaces {
		if iface.Name == pfInterface.If {
			result.Physical = iface
		}
	}
	return result
}

func populateInterfaces(p []inform.Interface, pfsense []pfconf.Interface, translation PfSenseTranslationTable) PfSenseTranslation {
	result := PfSenseTranslation{}
	for _, iface := range pfsense {
		switch iface.XMLName.Local {
		case translation.Lan:
			result.Lan = prepareRealInterface(iface, p, "eth1", "lan")
		case translation.Wan:
			result.Wan = prepareRealInterface(iface, p, "eth0", "wan")
		case translation.Wan2:
			result.Wan2 = prepareRealInterface(iface, p, "eth2", "wan2")
		case translation.Uid:
			result.Uid = prepareRealInterface(iface, p, "", "uid")
		}

	}
	return result
}

func computeMacFromIp(ip string) inform.HardwareAddr {
	ipObject := net.ParseIP(ip)
	if ipObject == nil {
		return nil
	}
	hwAddress := inform.HardwareAddr{0xBE, 0xEF}
	return append(hwAddress, ipObject[len(ipObject)-4:]...)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetPfSenseVersion() string {
	if fileExists("/etc/version") {
		contentVersion, err := ioutil.ReadFile("/etc/version")
		if err != nil {
			return ""
		}
		version := strings.TrimSpace(string(contentVersion))
		if len(version) == 0 {
			return ""
		}
		if fileExists("/etc/version.patch") {
			contentPatch, err := ioutil.ReadFile("/etc/version.patch")
			if err != nil {
				return ""
			}
			patch := strings.TrimSpace(string(contentPatch))
			if len(patch) > 0 {
				version = version + "-p" + patch
			}
		}
		return version
	}
	return ""
}

func RequestFromPfsense(address string, version string, pfsense pfconf.Configuration, translation PfSenseTranslationTable, speedtest *inform.SpeedTestStatus) (inform.Inform, error) {
	now := time.Now()
	request := inform.Inform{
		ConfigNetworkWan:     inform.NetworkConfig{Type: inform.NetworkConfigDhcp},
		ConfigNetworkWan2:    inform.NetworkConfig{Type: inform.NetworkConfigDisabled},
		BootRomVersion:       "unifi-v1.5.2.206-g44e4c8bc",
		ConfigVersion:        version,
		Default:              len(version) == 0 || version == "0123456789abcdef",
		DiscoveryResponse:    false,
		FirmwareCapabilities: int32(^uint32(0) >> 1),
		GuestToken:           "",
		HasEth1:              false,
		HasSshDisable:        true,
		Hostname:             "",
		IntfTable:            nil,
		InformUrl:            address,
		Ip:                   "",
		LastError:            "",
		Locating:             false,
		Mac:                  nil,
		Model:                "",
		ModelDisplay:         "",
		Netmask:              "",
		QrId:                 "",
		RequiredVersion:      "0.0.1",
		SelfrunBeacon:        true,
		Serial:               "",
		SpectrumScanning:     false,
		State:                2,
		StreamToken:          "",
		SysStats:             inform.SysStats{},
		Time:                 now.Unix(),
		Uplink:               "",
		Uptime:               0,
		Version:              "",
		PortTable:            []inform.Port{},
	}

	if speedtest != nil {
		request.SpeedtestStatus = speedtest
	}

	// Get controller IP
	request.InformIp = getIpFromURL(address)

	// Host infos
	host, err := hoststats.Info()
	if err == nil {
		request.Uptime = host.Uptime
		request.Hostname = host.Hostname
		request.Model = "UGW3"
		//request.ModelDisplay = strings.Title(host.Platform)
		request.ModelDisplay = "UniFi-Gateway-3"
		request.Version = host.PlatformVersion
		pfsenseVersion := GetPfSenseVersion()
		if len(pfsenseVersion) > 0 {
			request.Version = pfsenseVersion
		}
		request.BootRomVersion = host.PlatformVersion
	}

	// System statistics
	sys, err := SysStats()
	if err == nil {
		request.SysStats = sys
	}

	// Interfaces
	ifaces, err := Network()
	if err == nil {
		request.IntfTable = make([]inform.Interface, 0)
		request.PortTable = make([]inform.Port, 0)
		table := populateInterfaces(ifaces, pfsense.Interfaces.List, translation)
		if table.Wan.Pfsense.If != "" {
			// General
			request.Uplink = table.Wan.UnifiName
			request.Mac = table.Wan.Physical.Mac
			request.Serial = table.Wan.Physical.Mac.HexString()
			request.Ip = table.Wan.Physical.Ip
			request.Netmask = table.Wan.Physical.Netmask

			wan := table.Wan.Physical
			wan.Name = table.Wan.UnifiName
			gateway, err := Gateway()
			if err == nil {
				wan.Gateways = append(wan.Gateways, gateway.String())
			}
			if len(pfsense.System.DnsServers) > 0 {
				wan.Nameservers = pfsense.System.DnsServers
			}
			if speedtest != nil {
				wan.Latency = speedtest.Latency
			}
			request.IntfTable = append(request.IntfTable, wan)
			request.PortTable = append(request.PortTable, inform.Port{
				IfName: wan.Name,
				Name:   "wan",
			})
			if table.Wan.Pfsense.Ip != "dhcp" {
				request.ConfigNetworkWan = inform.NetworkConfig{
					Type:    "static",
					Ip:      table.Wan.Physical.Ip,
					Netmask: table.Wan.Physical.Netmask,
					Gateway: table.Wan.Pfsense.Gateway,
					IfName:  table.Wan.UnifiName,
				}
				for i, ns := range pfsense.System.DnsServers {
					switch i {
					case 1:
						request.ConfigNetworkWan.Dns1 = ns
						continue
					case 2:
						request.ConfigNetworkWan.Dns2 = ns
						continue
					}
					break
				}
			}
		}
		if table.Wan2.Pfsense.If != "" {
			wan := table.Wan2.Physical
			wan.Name = table.Wan2.UnifiName
			request.IntfTable = append(request.IntfTable, wan)
			request.PortTable = append(request.PortTable, inform.Port{
				IfName: wan.Name,
				Name:   "wan2",
			})
			if table.Wan2.Pfsense.Ip != "dhcp" {
				request.ConfigNetworkWan2 = inform.NetworkConfig{
					Type:    "static",
					Ip:      table.Wan2.Physical.Ip,
					Netmask: table.Wan2.Physical.Netmask,
					Gateway: table.Wan2.Pfsense.Gateway,
					IfName:  table.Wan2.UnifiName,
				}
				for i, ns := range pfsense.System.DnsServers {
					switch i {
					case 1:
						request.ConfigNetworkWan2.Dns1 = ns
						continue
					case 2:
						request.ConfigNetworkWan2.Dns2 = ns
						continue
					}
					break
				}
			}
			//
		}
		if table.Lan.Pfsense.If != "" {
			lan := table.Lan.Physical
			lan.Name = table.Lan.UnifiName
			request.IntfTable = append(request.IntfTable, lan)
			request.PortTable = append(request.PortTable, inform.Port{
				IfName: lan.Name,
				Name:   "lan",
			})
		}
		if table.Uid.Physical.Ip != "" {
			mac := computeMacFromIp(table.Uid.Physical.Ip)
			request.Mac = mac
			request.Serial = mac.HexString()
		}
	}

	return request, err
}
