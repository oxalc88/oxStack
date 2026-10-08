package main

import (
	"fmt"
	"os"
)

// ANSI colors
const (
	green  = "\033[0;32m"
	yellow = "\033[0;33m"
	red    = "\033[0;31m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

func infof(format string, a ...any)  { fmt.Printf(green+"+"+reset+" "+format+"\n", a...) }
func warnf(format string, a ...any)  { fmt.Printf(yellow+"!"+reset+" "+format+"\n", a...) }
func errorf(format string, a ...any) { fmt.Fprintf(os.Stderr, red+"x"+reset+" "+format+"\n", a...) }

func main() {
	if len(os.Args) < 2 {
		cmdHelp()
		return
	}

	switch os.Args[1] {
	case "install":
		cmdInstall()
	case "catalog":
		if err := cmdCatalog(); err != nil {
			errorf("%v", err)
			os.Exit(1)
		}
	case "sync":
		cmdSync()
	case "update":
		cmdUpdate()
	case "pull-config":
		cmdPullConfig()
	case "uninstall":
		cmdUninstall()
	case "help", "--help", "-h":
		cmdHelp()
	default:
		errorf("Unknown command: %s", os.Args[1])
		fmt.Println()
		cmdHelp()
		os.Exit(1)
	}
}

func cmdHelp() {
	fmt.Printf(bold + "oxstack" + reset + " — oxStack CLI\n")
	fmt.Println()
	fmt.Println("Usage: oxstack <command> [options]")
	fmt.Println("  install --global|--project DIR|--profile NAME [--selection FILE] [--runtime codex|claude|opencode] [--dry-run]")
	fmt.Println("  uninstall --global|--project DIR|--profile NAME [--runtime ...] [--dry-run]")
	fmt.Println("  catalog import NAME --source DIR   Import a reviewed skill without activation")
	fmt.Println("  sync / update                     Maintain forked gstack catalog sources")
	fmt.Println("  help                              Show this help")

	fmt.Printf("Repo: %s\n", repoRoot())
}

// run executes a command, printing its combined output. Returns the output and any error.
func run(name string, args ...string) (string, error) {
	return runDir("", name, args...)
}

// runDir executes a command in the given directory.
func runDir(dir, name string, args ...string) (string, error) {
	cmd := execCommand(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// runSilent executes a command, suppressing output. Returns the output and any error.
func runSilent(name string, args ...string) (string, error) {
	return runSilentDir("", name, args...)
}

// runSilentDir executes a command silently in the given directory.
func runSilentDir(dir, name string, args ...string) (string, error) {
	cmd := execCommand(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// execCommand creates an exec.Cmd — extracted so it's easy to find the import.
func execCommand(name string, args ...string) *execCmd {
	return newExecCmd(name, args...)
}
