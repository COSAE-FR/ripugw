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
	"bytes"
	"encoding/hex"
	"fmt"
)

// Key management
type Key []byte

func (k Key) IsValid() bool {
	return len(k) == len(DefaultKey)
}

func (k Key) IsDefault() bool {
	return bytes.Equal(k, DefaultKey)
}

func (k Key) String() string {
	return hex.EncodeToString(k)
}

func KeyFromString(keyString string) (Key, error) {
	keyBytes, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, err
	}
	key := Key(keyBytes)
	if !key.IsValid() {
		return nil, fmt.Errorf("invalid key length")
	}
	return key, nil
}

var (
	DefaultKey = Key([]byte{
		0xba, 0x86, 0xf2, 0xbb,
		0xe1, 0x07, 0xc7, 0xc5,
		0x7e, 0xb5, 0xf2, 0x69,
		0x07, 0x75, 0xc7, 0x12,
	})
)
