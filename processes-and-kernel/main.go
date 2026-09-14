package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	f := flag.String("flag", "", "section of the program to run")

	flag.Parse()

	switch *f {
	case "processes":
		res := childProcesses("ls", []string{"-l", "/home/"})
		fmt.Printf("[processes] response: %s\n", res)
	case "signals":
		unixSignals()
	case "fds":
		fileDescriptors()
	case "sds":
		sockets()
	case "cpu":
		fmt.Printf("fibonacci process PID: %d\n", os.Getpid())
		_ = fibonacci(52)
	default:
		fmt.Fprint(os.Stderr, "valid options: processes, signals, fds, sds, cpu\n")
		os.Exit(1)
	}
}
