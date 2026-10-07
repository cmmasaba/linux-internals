package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	stepGB := flag.Float64("step", 1, "GB to allocate at each interval")
	interval := flag.Duration("interval", 2*time.Second, "time between allocations")
	maxGB := flag.Float64("max", 0, "max memory to be allocated, 0 is unlimited")
	flag.Parse()

	const (
		GB       = 1024 * 1024 * 1024
		pageSize = 4096
	)

	stepBytes := int(*stepGB * GB)

	var (
		blocks    [][]byte
		allocated uint64
	)

	fmt.Printf("Process PID: %d\n", os.Getpid())
	fmt.Printf("Allocating %.2f GB of memory every %d\n", *stepGB, *interval)

	if *maxGB == 0 {
		fmt.Println("Maximum memory unlimited")
	} else {
		fmt.Printf("Maximum memory: %.2f GB\n", *maxGB)
	}

	for {
		block := make([]byte, stepBytes)

		// touch every page to force the kernel to back it with physical RAM(demand paging)
		for i := 0; i < len(block); i += pageSize {
			block[i] = 1
		}

		// keep reference to the page so it is not cleaned up by the GC
		blocks = append(blocks, block)

		allocated += uint64(stepBytes)

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		fmt.Printf("allocated: %.2f GB | Go heap size: %.2f GB | System total: %.2f GB\n",
			float64(allocated)/GB, float64(m.HeapAlloc)/GB, float64(m.Sys)/GB,
		)

		if *maxGB > 0 && float64(allocated)/GB >= *maxGB {
			fmt.Println("Maximum reached")
			signals := make(chan os.Signal, 1)

			signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

			<-signals

			runtime.KeepAlive(blocks)

			fmt.Println("Exiting...")

			return
		}

		time.Sleep(*interval)
	}
}
