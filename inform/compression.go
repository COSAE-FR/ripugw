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
	"compress/zlib"
	"github.com/golang/snappy"
	"io/ioutil"
)

func CompressZLib(data []byte) (result []byte, err error) {
	var b bytes.Buffer

	w := zlib.NewWriter(&b)
	if _, err = w.Write(data); err != nil {
		return
	}
	if err = w.Close(); err != nil {
		return
	}

	return b.Bytes(), err
}

func DecompressZLib(data []byte) (result []byte, err error) {
	b := bytes.NewReader(data)

	r, err := zlib.NewReader(b)
	if err != nil {
		return
	}

	result, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}

	return result, r.Close()
}

func CompressSnappy(data []byte) ([]byte, error) {
	return snappy.Encode(nil, data), nil
}

func DecompressSnappy(data []byte) ([]byte, error) {
	return snappy.Decode(nil, data)
}
