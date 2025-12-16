package util

import (
	"fmt"
	"runtime"
	"testing"
)

func TestUtil1(t *testing.T) {
	wNum := runtime.NumCPU()
	fmt.Printf("wNum: %v\n", wNum)
}
