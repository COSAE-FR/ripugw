/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package collect

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

func getInterfaceMediaForOs(iface string) (InterfaceMedia, error) {
	base := path.Join("/sys/class/net", iface)
	speedFile := path.Join(base, "speed")
	duplexFile := path.Join(base, "duplex")
	media := InterfaceMedia{
		Speed:      1000,
		FullDuplex: true,
	}
	if _, err := os.Stat(base); err != nil {
		return media, err
	}
	if _, err := os.Stat(speedFile); err != nil {
		return media, err
	}
	speedHandle, err := os.Open(speedFile)
	if err != nil {
		return media, err
	}
	speedString, err := ioutil.ReadAll(speedHandle)
	if err != nil {
		return media, err
	}
	speed, err := strconv.Atoi(string(speedString))
	if err != nil {
		return media, err
	}
	media.Speed = uint64(speed)
	if _, err := os.Stat(duplexFile); err != nil {
		return media, err
	}
	duplexHandle, err := os.Open(duplexFile)
	if err != nil {
		return media, err
	}
	duplex, err := ioutil.ReadAll(duplexHandle)
	if err != nil {
		return media, err
	}
	media.FullDuplex = string(duplex) == "full"
	return media, err
}
