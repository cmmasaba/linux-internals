package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"
)

func filesAndSockets() {
	f, err := os.Open("test_file")
	if err != nil {
		log.Fatalf("error opening file: %s", err)
	}

	log.Printf("PID of the process: %d", os.Getpid())

	// sleep for 2 mins to allow running shell commands
	time.Sleep(time.Minute * 2)

	f.Close()

	log.Print("closing file descriptor")

	time.Sleep(time.Minute * 1)
}

func leakFileDescriptors() {
	ok := true
	var fds []*os.File

	for ok {
		f, err := os.Open("test_file")
		if err != nil {
			log.Printf("error opening file: %s", err)
			ok = false
		}

		fds = append(fds, f)
	}

	log.Printf("Process PID: %d", os.Getpid())

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
		log.Fatalf("error opening server: %s", err)
	}

	defer l.Close()

	for {
		select {
		case <-ctx.Done():
			log.Printf("closing server...")
			return
		default:
		}

		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("error accepting connection: %s", err)
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
		log.Fatalf("error dialing connection: %s", err)
	}

	defer conn.Close()

	b := make([]byte, 1024)

	_, err = conn.Read(b)
	if err != nil {
		log.Fatalf("error reading from socket: %s", err)
	}
}

func sockets() {
	start := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*2))
	defer cancel()

	go server(ctx)

	for range 100 {
		go client(ctx)
	}

	<-ctx.Done()

	log.Printf("total time: %f", time.Since(start).Minutes())
}
