package main

import (
	"log"
	"os"
	"time"
)

func main() {
	childProcesses("ls", []string{"-l", "/home/"})

	unixSignals()

	filesAndSockets()

	leakFileDescriptors()

	sockets()

	var n uint = 55
	now := time.Now()
	log.Printf("fibonacci process PID: %d", os.Getpid())
	res := fibonacci(n)
	log.Printf("duration: %.1f mins", time.Since(now).Minutes())
	log.Printf("fibonacci of %d is %d", n, res)
}
