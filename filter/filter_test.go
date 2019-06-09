package filter

import (
	"fmt"
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	filter := New(1)
	filter.Check("1")
	time.Sleep(1e8)
	filter.Check("2")
	time.Sleep(1e8)
	filter.Check("3")
	filter.print()
	if filter.Check("1") {
		t.Error("value check failed")
	}
	fmt.Println("")
	time.Sleep(2e9)
	if !filter.Check("1") {
		t.Error("value check failed")
	}
	filter.print()
}
