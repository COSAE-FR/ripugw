/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package collect

import (
	"github.com/jackpal/gateway"
	"net"
)

func getGateway() (net.IP, error) {
	return gateway.DiscoverGateway()
}
