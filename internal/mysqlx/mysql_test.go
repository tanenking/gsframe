package mysqlx

import (
	"fmt"
	"strings"
	"testing"
)

func TestTrim(t *testing.T) {
	s := `
	select  s
	`

	fmt.Println(s)
	s1 := strings.TrimSpace(s)
	fmt.Println(s1)
}
