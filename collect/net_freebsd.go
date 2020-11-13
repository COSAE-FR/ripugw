/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package collect

import (
	"fmt"
	"golang.org/x/net/route"
	"net"
)

var defaultRoute = [4]byte{0, 0, 0, 0}

func getGateway() (net.IP, error) {
	rib, err := route.FetchRIB(0, route.RIBTypeRoute, 0)
	if err != nil {
		return nil, err
	}

	messages, err := route.ParseRIB(route.RIBTypeRoute, rib)
	if err != nil {
		return nil, err
	}

	for _, message := range messages {
		routeMessage := message.(*route.RouteMessage)
		addresses := routeMessage.Addrs

		var destination, gateway *route.Inet4Addr
		ok := false

		if destination, ok = addresses[0].(*route.Inet4Addr); !ok {
			continue
		}

		if gateway, ok = addresses[1].(*route.Inet4Addr); !ok {
			continue
		}

		if destination == nil || gateway == nil {
			continue
		}

		if destination.IP == defaultRoute {
			return net.IPv4(gateway.IP[0], gateway.IP[1], gateway.IP[2], gateway.IP[3]), nil
		}
	}
	return nil, fmt.Errorf("Cannot find gateway")
}
