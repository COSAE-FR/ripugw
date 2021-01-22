/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package pfconf

import (
	"encoding/xml"
	"strings"
)

type Configuration struct {
	XMLName     xml.Name   `xml:"pfsense"`
	FileVersion string     `xml:"version"`
	System      System     `xml:"system"`
	Interfaces  Interfaces `xml:"interfaces"`
	Routes      []Route    `xml:"staticroutes>route"`
	Syslog      Syslog     `xml:"syslog"`
	Revision    Revision   `xml:"revision"`
	Gateways    []Gateway  `xml:"gateways>gateway_item"`
	GatewayIpv4 string     `xml:"gateways>defaultgw4"`
	GatewayIpv6 string     `xml:"gateways>defaultgw6"`
	SysCtls     []SysCtl   `xml:"sysctl>item"`
}

func (c *Configuration) Finalize() error {
	err := c.System.Finalize()
	return err
}

type System struct {
	XMLName          xml.Name `xml:"system"`
	Hostname         string   `xml:"hostname"`
	Timezone         string   `xml:"timezone"`
	Domain           string   `xml:"domain"`
	DnsServersString string   `xml:"dnsserver"`
	DnsServers       []string `xml:"-"`
	Timeservers      string   `xml:"timeservers"`
}

func (s *System) Finalize() error {
	var err error
	if len(s.DnsServersString) > 0 {
		s.DnsServers = strings.Split(s.DnsServersString, ",")
	}
	return err
}

type Interfaces struct {
	List []Interface `xml:",any"`
}

type BoolIfElementPresent bool

type Interface struct {
	XMLName     xml.Name
	If          string               `xml:"if"`
	Enable      BoolIfElementPresent `xml:"enable"`
	BlockBogons BoolIfElementPresent `xml:"blockbogons"`
	SpoofMac    BoolIfElementPresent `xml:"spoofmac"`
	Description string               `xml:"descr"`
	Ip          string               `xml:"ipaddr"`
	Subnet      uint8                `xml:"subnet"`
	Gateway     string               `xml:"gateway"`
}

type Route struct {
	Network     string `xml:"network"`
	Gateway     string `xml:"gateway"`
	Description string `xml:"descr"`
}

type Syslog struct {
	FilterDescriptions string               `xml:"filterdescriptions"`
	Nentries           string               `xml:"nentries"`
	Remoteserver       string               `xml:"remoteserver"`
	Remoteserver2      string               `xml:"remoteserver2"`
	Remoteserver3      string               `xml:"remoteserver3"`
	SourceIp           string               `xml:"sourceip"`
	Protocol           string               `xml:"ipproto"`
	LogAll             BoolIfElementPresent `xml:"logall"`
	Enable             BoolIfElementPresent `xml:"enable"`
}

type Revision struct {
	Time        string `xml:"time"`
	Description string `xml:"description"`
	Username    string `xml:"username"`
}

type Gateway struct {
	Interface      string               `xml:"interface"`
	Gateway        string               `xml:"gateway"`
	Name           string               `xml:"name"`
	Weight         string               `xml:"weight"`
	Protocol       string               `xml:"ipprotocol"`
	Description    string               `xml:"descr"`
	MonitorDisable BoolIfElementPresent `xml:"monitor_disable"`
	ActionDisable  BoolIfElementPresent `xml:"action_disable"`
}

type SysCtl struct {
	Name        string `xml:"tunable"`
	Value       string `xml:"value"`
	Description string `xml:"descr"`
}
