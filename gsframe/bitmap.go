package gsframe

import (
	"errors"
	"math"
	"unsafe"
)

type BitMap interface {
	Set(idx uint32)
	Delete(idx uint32)
	Test(idx uint32) (bool, error)
	ToJson() string
	Full() bool
}

type bitmap struct {
	Bits      []uint64
	Count     uint32
	BitsSize  uint32
	BitsCount uint32
	BitsTotal uint32
}

func NewBitMap(bitssize uint32) *bitmap {
	temp := uint64(0)
	size := uint32(unsafe.Sizeof(temp) * 8)
	count := uint32(math.Ceil(float64(bitssize) / float64(size)))
	bf := &bitmap{
		Bits:      make([]uint64, count),
		Count:     count,
		BitsSize:  size,
		BitsCount: 0,
		BitsTotal: bitssize,
	}
	return bf
}

func NewBloomFilterFromJsonData(bytes string) *bitmap {
	bf := &bitmap{}
	if err := FromJson(bytes, bf); err != nil {
		return nil
	}
	return bf
}

func (bf *bitmap) Set(idx uint32) {
	index := idx >> 6
	if index >= bf.Count {
		return
	}
	value := idx & 63
	if value >= bf.BitsSize {
		return
	}
	exists := (bf.Bits[index] & (1 << value)) > 0
	if exists {
		bf.BitsCount++
	}
	bf.Bits[index] |= 1 << value
}
func (bf *bitmap) Delete(idx uint32) {
	index := idx >> 6
	if index >= bf.Count {
		return
	}
	value := idx & 63
	if value >= bf.BitsSize {
		return
	}
	exists := (bf.Bits[index] & (1 << value)) > 0
	if exists {
		bf.BitsCount--
	}
	bf.Bits[index] ^= 1 << value
}
func (bf *bitmap) Test(idx uint32) (bool, error) {
	index := idx >> 6
	if index >= bf.Count {
		return false, errors.New("index error")
	}
	value := idx & 63
	if value >= bf.BitsSize {
		return false, errors.New("value error")
	}
	exists := (bf.Bits[index] & (1 << value)) > 0
	return exists, nil
}

func (bf *bitmap) ToJson() string {
	return ToJson(bf)
}

func (bf *bitmap) Full() bool {
	return bf.BitsCount >= bf.BitsTotal
}
