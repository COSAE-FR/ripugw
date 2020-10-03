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
	"encoding/hex"
	"fmt"
)

type HardwareAddr []byte

func (m HardwareAddr) IsValid() bool {
	return len(m) == 6
}

func (m HardwareAddr) String() string {
	if !m.IsValid() {
		return ""
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		m[0], m[1], m[2], m[3], m[4], m[5])
}

func (m HardwareAddr) HexString() string {
	return hex.EncodeToString(m)
}

func (m HardwareAddr) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}
