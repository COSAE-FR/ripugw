/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package main

const defaultConfigFile = "/etc/ripugw/gateway.toml"

type ServiceConfig struct {
	File string `usage:"Gateway configuration file" default:"/etc/ripugw/gateway.toml"`
	Json bool   `usage:"Use JSON configuration file, not TOML"  default:"false"`
}
