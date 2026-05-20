package gsframe

import (
	"errors"
	"math"
	"unsafe"
)

type BloomFilter[T uint | uint8 | uint16 | uint32 | uint64] interface {
	Add(flag T)
	Exists(flag T) (bool, error)
	ToJson() string
	Full() bool
}

type bloomFilter[T uint | uint8 | uint16 | uint32 | uint64] struct {
	Flags          []T
	Count          T
	Flag_bits_size T
	FlagCount      T
	FlagTotal      T
}

func NewBloomFilter[T uint | uint8 | uint16 | uint32 | uint64](bitssize int) *bloomFilter[T] {
	temp := T(0)
	size := T(unsafe.Sizeof(temp) * 8)
	count := int(math.Ceil(float64(bitssize) / float64(size)))
	bf := &bloomFilter[T]{
		Flags:          make([]T, count),
		Count:          T(count),
		Flag_bits_size: size,
		FlagCount:      0,
		FlagTotal:      T(bitssize),
	}
	return bf
}

func NewBloomFilterFromJsonData[T uint | uint8 | uint16 | uint32 | uint64](bytes string) *bloomFilter[T] {
	bf := &bloomFilter[T]{}
	if err := FromJson(bytes, bf); err != nil {
		return nil
	}
	return bf
}

func (bf *bloomFilter[T]) getIndex(flag T) (arr_index, idx T) {
	arr_index = flag / bf.Flag_bits_size
	idx = flag % bf.Flag_bits_size
	return
}

func (bf *bloomFilter[T]) Add(flag T) {
	arr_index, idx := bf.getIndex(flag)
	if arr_index >= bf.Count {
		return
	}
	if idx >= bf.Flag_bits_size {
		return
	}
	n := T(1) << idx
	fg := bf.Flags[arr_index]
	bit := (n & fg)
	exists := bit > 0
	if !exists {
		bf.FlagCount++
		bf.Flags[arr_index] |= n
	}
}

func (bf *bloomFilter[T]) Exists(flag T) (bool, error) {
	arr_index, idx := bf.getIndex(flag)
	if arr_index >= bf.Count {
		return false, errors.New("arr index error")
	}
	if idx >= bf.Flag_bits_size {
		return false, errors.New("idx error")
	}
	n := T(1) << idx
	fg := bf.Flags[arr_index]
	bit := (n & fg)

	exists := bit > 0
	return exists, nil
}

func (bf *bloomFilter[T]) ToJson() string {
	return ToJson(bf)
}

func (bf *bloomFilter[T]) Full() bool {
	return bf.FlagCount >= bf.FlagTotal
}
