package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type commandOptions struct {
	Target        Target
	SelectionPath string
	DryRun        bool
}

func parseOptions(command string, args []string, root string) (commandOptions, error) {
	o := commandOptions{}
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	f.SetOutput(io.Discard)
	global := f.Bool("global", false, "explicit global scope")
	f.StringVar(&o.Target.Project, "project", "", "project directory (worktree supported)")
	f.StringVar(&o.Target.Profile, "profile", "", "on-demand profile name")
	f.StringVar(&o.Target.Runtime, "runtime", "codex", "codex, claude or opencode")
	f.StringVar(&o.SelectionPath, "selection", "", "selection TOML path")
	f.BoolVar(&o.DryRun, "dry-run", false, "preview without writing")
	if err := f.Parse(args); err != nil {
		return o, err
	}
	if f.NArg() > 0 {
		return o, fmt.Errorf("unexpected positional arguments")
	}
	count := 0
	if *global {
		count++
		o.Target.Scope = "global"
	}
	if o.Target.Project != "" {
		count++
		o.Target.Scope = "project"
		p, err := filepath.Abs(o.Target.Project)
		if err != nil {
			return o, err
		}
		if r, e := filepath.EvalSymlinks(p); e == nil {
			p = r
		}
		o.Target.Project = p
	}
	if o.Target.Profile != "" {
		count++
		o.Target.Scope = "profile"
	}
	if count != 1 {
		return o, fmt.Errorf("select exactly one scope: --global, --project DIR, or --profile NAME")
	}
	if err := o.Target.validate(); err != nil {
		return o, err
	}
	if command == "install" && o.SelectionPath == "" {
		switch o.Target.Scope {
		case "global":
			o.SelectionPath = filepath.Join(root, "selections", "global.toml")
		case "profile":
			o.SelectionPath = filepath.Join(root, "selections", "profiles", o.Target.Profile+".toml")
		default:
			return o, fmt.Errorf("project install requires --selection FILE")
		}
	}
	return o, nil
}
func runActivation(command string, args []string) error {
	e, err := newEnvironment(repoRoot())
	if err != nil {
		return err
	}
	o, err := parseOptions(command, args, e.Root)
	if err != nil {
		return err
	}
	var p *Plan
	if command == "install" {
		var s Selection
		s, err = readSelection(o.SelectionPath)
		if err != nil {
			return err
		}
		p, err = planInstall(e, o.Target, s)
	} else {
		p, err = planUninstall(e, o.Target)
	}
	if err != nil {
		return err
	}
	fmt.Printf("%s scope=%s runtime=%s\n", command, o.Target.Scope, o.Target.Runtime)
	for _, a := range p.Actions {
		fmt.Println(a)
	}
	if o.DryRun {
		fmt.Println("Dry run: no files changed.")
		return nil
	}
	if err = p.Apply(); err != nil {
		return err
	}
	fmt.Printf("Ownership: %s\n", o.Target.manifestPath(e))
	return nil
}
func confirm(in io.Reader, out io.Writer) bool {
	fmt.Fprint(out, "Type MIGRATE to modify the previewed legacy activation: ")
	s, _ := bufio.NewReader(in).ReadString('\n')
	return strings.TrimSpace(s) == "MIGRATE"
}
