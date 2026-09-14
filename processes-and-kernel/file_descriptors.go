package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"
)

func fileDescriptors() {
	ok := true
	var fds []*os.File

	for ok {
		f, err := os.Open("test_file")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening file: %s\n", err)
			ok = false
		}

		fds = append(fds, f)
	}

	fmt.Printf("Process PID: %d\n", os.Getpid())
	fmt.Printf("Number of open file descriptors: %d\n", len(fds))

	time.Sleep(time.Minute * 2)

	for _, f := range fds {
		f.Close()
	}
}

// count open fds for a process:
// ls -la /proc/3854255/fd | wc -l
//
// Find process with most open fds:
// for pid in /proc/[0-9]*;do                                                                                                                          git:main*
// echo "$(ls -l $pid/fd 2>/dev/null | wc -l) $(cat $pid/cmdline 2>/dev/null)"
// done | sort -rn | head -n 10
//
// See exactly what those fds point to
// lsof -p PID

func server(ctx context.Context) {
	l, err := net.Listen("tcp", ":8586")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening server: %s\n", err)
		os.Exit(1)
	}

	defer l.Close()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("closing server...")
			return
		default:
		}

		conn, err := l.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error accepting connection: %s\n", err)
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // failing to close conn leaks sockets file descriptors

	conn.Write([]byte("hello"))
}

func client(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}

	conn, err := net.Dial("tcp", "localhost:8586")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error dialing connection: %s\n", err)
		os.Exit(1)
	}

	defer conn.Close()

	b := make([]byte, 1024)

	_, err = conn.Read(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading from socket: %s\n", err)
		os.Exit(1)
	}
}

func sockets() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*2))
	defer cancel()

	fmt.Printf("Process PID: %d\n", os.Getpid())

	go server(ctx)

	for range 100 {
		go client(ctx)
	}

	<-ctx.Done()
}
