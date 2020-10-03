/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package collect

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
)

func getInterfaceMediaForOs(iface string) (InterfaceMedia, error) {
	re := regexp.MustCompile(`\tmedia:\s.+\s\((?P<speed>\d+)base.{2,}\s<(?P<duplex>[a-z]+)-duplex>\)\n`)

	media := InterfaceMedia{
		Speed:      1000,
		FullDuplex: true,
	}
	cmd := exec.Command("ifconfig", iface)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	match := re.FindSubmatch(out.Bytes())
	if len(match) == 3 {
		speed, err := strconv.Atoi(fmt.Sprintf("%s", match[1]))
		if err == nil {
			media.Speed = uint64(speed)
		}
		if fmt.Sprintf("%s", match[2]) == "half" {
			media.FullDuplex = false
		}
	}

	return media, nil
}
