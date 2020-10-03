/*
 * Copyright (c) 2017 ZAP Québec
 * Copyright (c) 2020 Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package binary

type Buffer []byte

func NewBuffer(length uint) Buffer {
	return make(Buffer, length)
}

func (b Buffer) Write(offset uint, bytes []byte) {
	copy(b[offset:], bytes)
}

func (b Buffer) WriteUInt16BE(offset uint, v uint16) {
	b[offset] = byte((v & 0xff00) >> 8)
	b[offset+1] = byte(v & 0x00ff)
}

func (b Buffer) WriteUInt32BE(offset uint, v uint32) {
	b[offset] = byte((v & 0xff000000) >> 24)
	b[offset+1] = byte((v & 0x00ff0000) >> 16)
	b[offset+2] = byte((v & 0x0000ff00) >> 8)
	b[offset+3] = byte(v & 0x000000ff)
}

func (b Buffer) Read(offset, end uint) []byte {
	return b[offset:end]
}

func (b Buffer) ReadUInt16BE(offset uint) uint16 {
	return (uint16(b[offset]) << 8) + uint16(b[offset+1])
}

func (b Buffer) ReadUInt32BE(offset uint) uint32 {
	return (uint32(b[offset]) << 24) +
		(uint32(b[offset+1]) << 16) +
		(uint32(b[offset+2]) << 8) +
		uint32(b[offset+3])
}
