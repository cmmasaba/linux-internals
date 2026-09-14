package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func childProcesses(command string, args []string) string {
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting cmd stdout pipe: %s\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error waiting for cmd to start: %s\n", err)
		os.Exit(1)
	}

	res, err := io.ReadAll(stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading output: %s\n", err)
		os.Exit(1)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "error waiting for cmd to terminate: %s", err)
		os.Exit(1)
	}

	return string(res)
}
