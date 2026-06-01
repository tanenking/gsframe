package gsframe

import (
	"fmt"
	"testing"
)

func TestBloom(t *testing.T) {
	bl := NewBloomFilter(10000 * 10000)
	bl.Add(100)
	bl.Add(22)
	fmt.Println(bl.Exists(22))
	fmt.Println(bl.Exists(23))
	fmt.Println(bl.Exists(100))
}
