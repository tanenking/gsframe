package gsframe

import (
	"fmt"
	"testing"
)

func TestBloom(t *testing.T) {
	bl := NewBitMap(10000 * 10000)
	bl.Set(100)
	bl.Set(22)
	fmt.Println(bl.Test(22))
	fmt.Println(bl.Test(23))
	fmt.Println(bl.Test(100))
	bl.Delete(100)
	fmt.Println(bl.Test(22))
	fmt.Println(bl.Test(23))
	fmt.Println(bl.Test(100))
}
