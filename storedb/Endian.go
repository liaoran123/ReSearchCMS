package storedb

import (
	"encoding/binary"
	"unsafe"
)

var EndianOrder = Endian()

// 判断大小端
func Endian() binary.ByteOrder {
	var i int = 0x1
	ptr := unsafe.Pointer(&i)
	b := *(*byte)(ptr)
	if b == 1 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}
