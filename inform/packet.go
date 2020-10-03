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
	"errors"
	"github.com/gcrahay/ripugw/inform/binary"
)

const (
	MagicNumber   uint32 = 1414414933
	InformVersion uint32 = 0
	DataVersion   uint32 = 1

	EncryptFlag uint16 = 1
	ZlibFlag    uint16 = 2
	SnappyFlag  uint16 = 4
	GcmFlag     uint16 = 8
)

var (
	NilIv IV = []byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
)

type Packet struct {
	ap    HardwareAddr
	flags uint16
	key   Key
	Msg   Message
	mode  int
}

func NewPacket(ap HardwareAddr, msg Message, k Key, mode int) *Packet {
	flags := SnappyFlag
	if k != nil {
		flags = flags | EncryptFlag
		if mode == GCM {
			flags = flags | GcmFlag
		}
	}
	return &Packet{
		ap:    ap,
		key:   k,
		mode:  mode,
		Msg:   msg,
		flags: flags,
	}
}

func (p Packet) IsEncrypted() bool {
	return p.flags&EncryptFlag == EncryptFlag
}

func (p Packet) IsGcmEncrypted() bool {
	return p.IsEncrypted() && p.flags&GcmFlag == GcmFlag
}

func (p Packet) IsZLib() bool {
	return p.flags&ZlibFlag == ZlibFlag
}

func (p Packet) IsSnappy() bool {
	return p.flags&SnappyFlag == SnappyFlag
}

func (p Packet) Marshal() (result []byte, err error) {
	msg := p.Msg.Marshal()

	if p.IsZLib() {
		msg, err = CompressZLib(msg)
		if err != nil {
			return
		}
	} else if p.IsSnappy() {
		msg, err = CompressSnappy(msg)
		if err != nil {
			return
		}
	}

	iv := NilIv
	if p.IsEncrypted() {
		iv, err = GenerateIV()
		if err != nil {
			return
		}
	}

	if len(p.ap) != 6 {
		return nil, errors.New("Invalid length of MAC address")
	}

	b := binary.NewBuffer(uint(40))
	b.WriteUInt32BE(0, MagicNumber)
	b.WriteUInt32BE(4, InformVersion)
	b.Write(8, p.ap)
	b.WriteUInt16BE(14, p.flags)
	b.Write(16, iv)
	b.WriteUInt32BE(32, DataVersion)
	if p.IsEncrypted() {
		if p.IsGcmEncrypted() {
			b.WriteUInt32BE(36, uint32(len(msg)+16))
		}
		msg, err = Encrypt(p.mode, iv, p.key, msg, b[:])
		if err != nil {
			return
		}
		if !p.IsGcmEncrypted() {
			b.WriteUInt32BE(36, uint32(len(msg)))
		}
	} else {
		b.WriteUInt32BE(36, uint32(len(msg)))
	}
	b = append(b, msg...)

	return binary.Buffer(b), nil
}

func (p *Packet) Unmarshal(data []byte, keyFetcher func(addr HardwareAddr) (Key, error)) (err error) {

	b := binary.Buffer(data)

	if len(data) < 40 {
		return errors.New("Invalid packet length.")
	}
	dataLength := uint(b.ReadUInt32BE(36))
	if uint(len(data)) < dataLength+40 {
		return errors.New("Invalid packet length.")
	}
	if b.ReadUInt32BE(0) != MagicNumber {
		return errors.New("Invalid magic number at start of packet.")
	}
	if b.ReadUInt32BE(4) != InformVersion {
		return errors.New("Unkwown inform version.")
	}
	if b.ReadUInt32BE(32) != DataVersion {
		return errors.New("Unkwown data version.")
	}

	p.ap = b.Read(8, 14)

	p.flags = b.ReadUInt16BE(14)

	msg := b.Read(40, 40+dataLength)
	if p.IsEncrypted() {
		iv := IV(b.Read(16, 32))
		p.key, err = keyFetcher(p.ap)
		if err != nil {
			return err
		}
		mode := CBC
		if p.IsGcmEncrypted() {
			mode = GCM
		}
		msg, err = Decrypt(mode, iv, p.key, msg, b[:40])
		if err != nil {
			return err
		}
	}

	if p.IsZLib() {
		msg, err = DecompressZLib(msg)
		if err != nil {
			return err
		}
	} else if p.IsSnappy() {
		msg, err = DecompressSnappy(msg)
		if err != nil {
			return err
		}
	}

	p.Msg, err = Unmarshal(msg)
	return err
}
