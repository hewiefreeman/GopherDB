package helpers

import (
	"fmt"
	"runtime"
	//"time"
)

func PrintMemUsage() {
	fmt.Println("/////////////////////////////////////////////////////////////////////")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
    return b / 1024 / 1024
}
