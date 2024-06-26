package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

	var coffeeOptions = []string{"-w", fmt.Sprint(cmd.Process.Pid)}
	if val, found := os.LookupEnv("COFFEE_OPTIONS"); found {
		coffeeOptions = append(coffeeOptions, strings.Split(val, ",")...)
	} else {
		coffeeOptions = append(coffeeOptions, "-dim")
	}

	coffee := exec.CommandContext(ctx, "/usr/bin/caffeinate", coffeeOptions...)
	go func() {
		defer cancel()

		if err := coffee.Run(); err != nil {
			if !errors.Is(ctx.Err(), context.Canceled) {
				fmt.Printf("failed to start caffeinate: %v\n", err)
			}

			// Do not let primary command run if we cannot track it
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
	}()

	if err := cmd.Wait(); err != nil {
		procState, _ := cmd.Process.Wait()
		os.Exit(procState.ExitCode())
	}
}
