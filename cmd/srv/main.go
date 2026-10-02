package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"srv.exe.dev/srv"
)

var flagListenAddr = flag.String("listen", ":8000", "address to listen on")

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run() error {
	flag.Parse()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	// Bind the port before opening the database: if another instance is
	// still running we must fail fast without touching (and locking) its DB.
	ln, err := net.Listen("tcp", *flagListenAddr)
	if err != nil {
		return err
	}
	server, err := srv.New("db.sqlite3", hostname)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}
	return server.ServeListener(ln)
}
