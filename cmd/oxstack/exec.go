package main

import "os/exec"

// execCmd wraps exec.Cmd so main.go doesn't import os/exec directly.
type execCmd = exec.Cmd

func newExecCmd(name string, args ...string) *execCmd {
	return exec.Command(name, args...)
}

// commandExists checks whether a command is available in PATH.
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
