package gsframe

import (
	"errors"
	"math"
	"unsafe"
)

type BloomFilter interface {
	Add(flag uint32)
	Exists(flag uint32) (bool, error)
	ToJson() string
	Full() bool
}

type bloomFilter struct {
	Flags          []uint64
	Count          uint32
	Flag_bits_size uint32
	FlagCount      uint32
	FlagTotal      uint32
}

func NewBloomFilter(bitssize uint32) *bloomFilter {
	temp := uint64(0)
	size := uint32(unsafe.Sizeof(temp) * 8)
	count := uint32(math.Ceil(float64(bitssize) / float64(size)))
	bf := &bloomFilter{
		Flags:          make([]uint64, count),
		Count:          count,
		Flag_bits_size: size,
		FlagCount:      0,
		FlagTotal:      bitssize,
	}
	return bf
}

func NewBloomFilterFromJsonData(bytes string) *bloomFilter {
	bf := &bloomFilter{}
	if err := FromJson(bytes, bf); err != nil {
		return nil
	}
	return bf
}

func (bf *bloomFilter) getIndex(flag uint32) (arr_index, idx uint32) {
	arr_index = flag / bf.Flag_bits_size
	idx = flag % bf.Flag_bits_size
	return
}

func (bf *bloomFilter) Add(flag uint32) {
	arr_index, idx := bf.getIndex(flag)
	if arr_index >= bf.Count {
		return
	}
	if idx >= bf.Flag_bits_size {
		return
	}
	n := uint64(1) << idx
	fg := bf.Flags[arr_index]
	bit := n & fg
	exists := bit > 0
	if !exists {
		bf.FlagCount++
		bf.Flags[arr_index] |= n
	}
}
func (bf *bloomFilter) Remove(flag uint32) {
	arr_index, idx := bf.getIndex(flag)
	if arr_index >= bf.Count {
		return
	}
	if idx >= bf.Flag_bits_size {
		return
	}
	n := uint64(1) << idx
	fg := bf.Flags[arr_index]
	bit := n & fg
	exists := bit > 0
	if exists {
		bf.FlagCount--
		bf.Flags[arr_index] ^= n
	}
}
func (bf *bloomFilter) Exists(flag uint32) (bool, error) {
	arr_index, idx := bf.getIndex(flag)
	if arr_index >= bf.Count {
		return false, errors.New("arr index error")
	}
	if idx >= bf.Flag_bits_size {
		return false, errors.New("idx error")
	}
	n := uint64(1) << idx
	fg := bf.Flags[arr_index]
	bit := n & fg

	exists := bit > 0
	return exists, nil
}
func (bf *bloomFilter) ToJson() string {
	return ToJson(bf)
}

func (bf *bloomFilter) Full() bool {
	return bf.FlagCount >= bf.FlagTotal
}
