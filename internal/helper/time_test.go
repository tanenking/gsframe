package helper_test

import (
	"fmt"
	"testing"

	"github.com/tanenking/gsframe/internal/timex"
)

func TestTimestamp(t *testing.T) {
	fmt.Println(timex.GetNowTimestamp())
}
