/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package collect

import (
	"errors"
	"fmt"
	"github.com/COSAE-FR/ripugw/inform"
	"github.com/jackpal/gateway"
	cpustats "github.com/shirou/gopsutil/cpu"
	hoststats "github.com/shirou/gopsutil/host"
	loadstats "github.com/shirou/gopsutil/load"
	memstats "github.com/shirou/gopsutil/mem"
	netstats "github.com/shirou/gopsutil/net"
	"net"
	"net/url"
	"strings"
	"time"
)

func Network() ([]inform.Interface, error) {
	counters, err := netstats.IOCounters(true)
	if err != nil {
		return nil, err
	}

	var data []inform.Interface
	for _, counter := range counters {
		unifiCounter := inform.Interface{
			Name:      counter.Name,
			TxBytes:   counter.BytesSent,
			RxBytes:   counter.BytesRecv,
			TxPackets: counter.PacketsSent,
			RxPackets: counter.PacketsRecv,
			TxErrors:  counter.Errout,
			RxErrors:  counter.Errin,
			TxDropped: counter.Dropout,
			RxDropped: counter.Dropin,
			Latency:   1, // Default Latency: without this, controller shows not Internet connection
		}
		unifiCounter.Drops = counter.Dropout + counter.Dropin

		ipAddress, err := GetIPForInterface(counter.Name)
		if err == nil {
			unifiCounter.Ip = ipAddress.IP.String()
			unifiCounter.Netmask = FormatMask(ipAddress.Mask)
		}

		iface, err := net.InterfaceByName(counter.Name)
		if err == nil {
			if iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			unifiCounter.Mac = inform.HardwareAddr(iface.HardwareAddr)
			unifiCounter.NumPort = iface.Index
			if iface.Flags&net.FlagUp != 0 {
				unifiCounter.Up = true
				unifiCounter.Enabled = true

				media, _ := GetInterfaceMedia(iface.Name)

				unifiCounter.FullDuplex = media.FullDuplex
				unifiCounter.Speed = media.Speed
			}
		}
		unifiCounter.Gateways = []string{}
		unifiCounter.Nameservers = []string{}
		data = append(data, unifiCounter)
	}
	return data, nil
}

func GetIPForInterface(interfaceName string) (ipAddress *net.IPNet, err error) {
	interfaces, _ := net.Interfaces()
	for _, inter := range interfaces {
		if inter.Name == interfaceName {
			if addrs, err := inter.Addrs(); err == nil {
				for _, addr := range addrs {
					switch ip := addr.(type) {
					case *net.IPNet:
						if ip.IP.To4() != nil {
							return ip, nil
						}
					}
				}
			}
		}
	}
	return ipAddress, errors.New("no IP found")
}

func FormatMask(mask net.IPMask) string {
	if len(mask) != 4 {
		return mask.String()
	}

	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}

func Gateway() (net.IP, error) {
	return gateway.DiscoverGateway()
}

func SysStats() (inform.SysStats, error) {

	load, err := loadstats.Avg()
	if err != nil {
		return inform.SysStats{}, err
	}
	data := inform.SysStats{
		LoadAvg1:  load.Load1,
		LoadAvg5:  load.Load5,
		LoadAvg15: load.Load15,
	}

	cpus, err := cpustats.Percent(0, false)
	if err == nil && len(cpus) > 0 {
		data.Cpu = uint64(cpus[0])
	}

	mem, err := memstats.VirtualMemory()
	if err != nil {
		return inform.SysStats{}, err
	}
	data.MemTotal = mem.Total
	data.MemUsed = mem.Used
	data.MemBuffer = mem.Buffers
	data.Mem = uint64(float64(mem.Used) / float64(mem.Total) * 100)

	return data, nil
}

func getIpFromURL(address string) string {
	ctrl, err := url.Parse(address)
	if err == nil {
		var addr net.IP
		host := ctrl.Host

		s := strings.Split(host, ":")
		if len(s) > 0 {
			host = s[0]
			addr = net.ParseIP(host)
		}

		if addr == nil {
			addrs, err := net.LookupIP(host)
			if err == nil && len(addrs) > 0 {
				addr = addrs[0]
			}
		}

		if addr != nil {
			return addr.String()
		}
		return host
	}
	return ""
}

func Request(address string, version string) (inform.Inform, error) {
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

	// Get controller IP
	request.InformIp = getIpFromURL(address)

	host, err := hoststats.Info()
	if err == nil {
		request.Uptime = host.Uptime
		request.Hostname = host.Hostname
		request.Model = "UGWXG"
		request.ModelDisplay = "UniFi Security Gateway XG-8"
		request.Version = "4.4.51.5287926 "
		request.BootRomVersion = host.PlatformVersion
	}

	ifaces, err := Network()
	if err == nil {
		request.IntfTable = ifaces
		request.EthernetTable = make([]inform.EthernetTableEntry, len(ifaces))
		for _, iface := range ifaces {
			request.EthernetTable = append(request.EthernetTable, inform.EthernetTableEntry{
				Name:    iface.Name,
				Mac:     iface.Mac.String(),
				NumPort: uint64(iface.NumPort),
			})
			if iface.Up && len(iface.Ip) > 0 {
				if request.Uplink == "" {
					request.Uplink = iface.Name
					request.Mac = iface.Mac
					request.Ip = iface.Ip
					request.Netmask = iface.Netmask
					request.Serial = iface.Mac.HexString()
					request.PortTable = append(request.PortTable, inform.Port{
						IfName: iface.Name,
						Name:   "wan",
					})
					request.ConfigNetworkWan.IfName = iface.Name
				} else {
					ifNumber := len(request.PortTable)
					ifString := fmt.Sprintf("%d", ifNumber)
					if ifNumber == 1 {
						ifString = ""
					}
					request.PortTable = append(request.PortTable, inform.Port{
						IfName: iface.Name,
						Name:   fmt.Sprintf("lan%s", ifString),
					})
				}
			}
		}
	}
	request.HasEth1 = len(ifaces) > 1

	sys, err := SysStats()
	if err == nil {
		request.SysStats = sys
	}

	return request, nil
}
