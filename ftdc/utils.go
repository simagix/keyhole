// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"encoding/binary"
	"io"
)

// GetUint32 -
func GetUint32(r io.Reader) uint32 {
	var size uint32
	binary.Read(r, binary.LittleEndian, &size)
	return size
}

// Uvarint -
func Uvarint(r io.ByteReader) uint64 {
	var x uint64
	var s uint
	var b byte
	var err error

	for i := 0; ; i++ {
		if b, err = r.ReadByte(); err != nil {
			return 0
		}
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0 // overflow
			}
			return x | uint64(b)<<s
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}
