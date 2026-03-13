package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/c0tton-fluff/sentinelone-mcp-server/config"
	"github.com/c0tton-fluff/sentinelone-mcp-server/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if _, err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	startWatchdog()

	s := server.NewMCPServer(
		"sentinelone",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	tools.Register(s)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// startWatchdog auto-exits when the parent process dies, preventing zombie MCP servers.
func startWatchdog() {
	ppid := os.Getppid()
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if err := syscall.Kill(ppid, 0); err != nil {
				os.Exit(0)
			}
		}
	}()
}
