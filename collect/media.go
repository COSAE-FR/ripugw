/*
 * Copyright (c) 2020. Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package collect

type InterfaceMedia struct {
	Speed      uint64
	FullDuplex bool
}

func GetInterfaceMedia(iface string) (InterfaceMedia, error) {
	return getInterfaceMediaForOs(iface)
}
