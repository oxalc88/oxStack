package main

import "os"

func cmdUninstall() {
	if err := runActivation("uninstall", os.Args[2:]); err != nil {
		errorf("%v", err)
		os.Exit(1)
	}
}
