package main

import (
	"io"
	"log"
	"os/exec"
)

func childProcesses(command string, args []string) string {
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("error getting cmd stdout pipe: %s\n", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("error waiting for cmd to start: %s\n", err)
	}

	res, err := io.ReadAll(stdout)
	if err != nil {
		log.Fatalf("error reading output: %s\n", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("error waiting for cmd to terminate: %s", err)
	}

	return string(res)
}
