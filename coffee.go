package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) == 1 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cmd *exec.Cmd
	if len(os.Args) == 2 {
		cmd = exec.CommandContext(ctx, os.Args[1])
	} else {
		cmd = exec.CommandContext(ctx, os.Args[1], os.Args[2:]...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	coffee := exec.CommandContext(ctx, "/usr/bin/caffeinate", "-d", "-w", fmt.Sprint(cmd.Process.Pid))
	defer func() {
		if err := coffee.Run(); err != nil {
			fmt.Printf("failed to start caffeinate: %v\n", err)
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
	}()

	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
