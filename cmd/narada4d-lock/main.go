package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	_ "github.com/powerman/narada4d/protocol/file"
	_ "github.com/powerman/narada4d/protocol/goose-postgres"
	_ "github.com/powerman/narada4d/protocol/mysql"
	"github.com/powerman/narada4d/schemaver"
)

func main() {
	log.SetFlags(0)

	schemaVer, err := schemaver.New()
	if err != nil {
		log.Fatalln("Failed to detect data schema version:", err)
	}

	os.Exit(lock(schemaVer, os.Args[1:]))
}

func lock(schemaVer *schemaver.SchemaVer, args []string) int {
	if os.Getenv(schemaver.EnvSkipLock) != "" {
		fmt.Println("Skip acquiring exclusive lock.")
		defer fmt.Println("Skip releasing exclusive lock.")
	} else {
		fmt.Println("Acquiring exclusive lock...")
		defer fmt.Println("Releasing exclusive lock...")
	}

	schemaVer.ExclusiveLock()
	defer schemaVer.Unlock()

	return run(args)
}

func run(args []string) (code int) {
	if len(args) == 0 {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		args = append(args, shell)
	}
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec // By design.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if wait, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				code = wait.ExitStatus()
			} else {
				code = 127
			}
		} else {
			log.Print(err)
		}
	}
	return code
}
