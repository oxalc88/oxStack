package main

import "os"

func cmdPullConfig() {
	errorf("pull-config is retired: edit an explicit selection; runtime toggles must not change the catalog")
	os.Exit(1)
}
