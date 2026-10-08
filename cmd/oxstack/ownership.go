package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CapabilityReference struct {
	Capability string `json:"capability"`
	Source     string `json:"source"`
	Version    string `json:"version"`
}
type Resource struct {
	References  []CapabilityReference `json:"references,omitempty"`
	Capability  string                `json:"capability"`
	Source      string                `json:"source"`
	Version     string                `json:"version"`
	Scope       string                `json:"scope"`
	Destination string                `json:"destination"`
	Runtime     string                `json:"runtime"`
	Kind        string                `json:"kind"`
	Key         []string              `json:"key,omitempty"`
	Digest      string                `json:"digest"`
	Value       any                   `json:"-"`
	Content     []byte                `json:"-"`
	Link        string                `json:"link,omitempty"`
}
type Manifest struct {
	Format    int        `json:"format"`
	Scope     string     `json:"scope"`
	Runtime   string     `json:"runtime"`
	Project   string     `json:"project,omitempty"`
	Profile   string     `json:"profile,omitempty"`
	Resources []Resource `json:"resources"`
}
type Snapshot struct {
	Path    string
	Content []byte
	Link    string
	Mode    os.FileMode
	Exists  bool
}
type Mutation struct{ Before, After Snapshot }
type Plan struct {
	Changes []Mutation
	Actions []string
}

func fingerprint(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func bytesDigest(b []byte) string  { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
func resourceID(r Resource) string { return r.Destination + "|" + r.Kind + "|" + fmt.Sprint(r.Key) }

func snapshot(path string) (Snapshot, error) {
	s := Snapshot{Path: path, Mode: 0600}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	s.Exists = true
	s.Mode = info.Mode().Perm()
	if info.Mode()&os.ModeSymlink != 0 {
		s.Link, err = os.Readlink(path)
		return s, err
	}
	if !info.Mode().IsRegular() {
		return s, fmt.Errorf("refuse non-file destination %s", path)
	}
	s.Content, err = os.ReadFile(path)
	return s, err
}
func safeParents(path string) error {
	for p := filepath.Dir(path); p != filepath.Dir(p); p = filepath.Dir(p) {
		s, err := os.Lstat(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if s.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refuse symlinked parent %s (resolve ownership first)", p)
		}
	}
	return nil
}
func sameSnapshot(a, b Snapshot) bool {
	return a.Exists == b.Exists && a.Link == b.Link && bytes.Equal(a.Content, b.Content) && (!a.Exists || a.Mode == b.Mode)
}
func writeSnapshot(s Snapshot) error {
	if err := safeParents(s.Path); err != nil {
		return err
	}
	if !s.Exists {
		err := os.Remove(s.Path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0700); err != nil {
		return err
	}
	if s.Link != "" {
		return os.Symlink(s.Link, s.Path)
	}
	f, err := os.CreateTemp(filepath.Dir(s.Path), ".oxstack-write-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(s.Mode); err == nil {
		_, err = f.Write(s.Content)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, s.Path)
}
func (p *Plan) Apply() error {
	// Preflight the entire plan before touching any destination. Recheck each
	// destination before replacing it, and roll back only our unchanged writes.
	for _, m := range p.Changes {
		if err := safeParents(m.Before.Path); err != nil {
			return err
		}
		s, err := snapshot(m.Before.Path)
		if err != nil {
			return err
		}
		if !sameSnapshot(s, m.Before) {
			return fmt.Errorf("destination changed since preview: %s", s.Path)
		}
	}
	done := []Mutation{}
	for _, m := range p.Changes {
		s, err := snapshot(m.Before.Path)
		if err == nil && !sameSnapshot(s, m.Before) {
			err = fmt.Errorf("destination changed: %s", s.Path)
		}
		if err == nil {
			err = writeSnapshot(m.After)
		}
		if err != nil {
			for i := len(done) - 1; i >= 0; i-- {
				x := done[i]
				now, e := snapshot(x.After.Path)
				if e == nil && sameSnapshot(now, x.After) {
					if x.After.Link != "" {
						e = os.Remove(x.After.Path)
					}
					if e == nil {
						e = writeSnapshot(x.Before)
					}
				} else if e == nil {
					e = fmt.Errorf("externally changed during rollback")
				}
				if e != nil {
					err = fmt.Errorf("%w; rollback %s: %v", err, x.Before.Path, e)
				}
			}
			return err
		}
		done = append(done, m)
	}
	return nil
}
func loadManifest(e Environment, t Target) (Manifest, error) {
	m := Manifest{Format: 1, Scope: t.Scope, Runtime: t.Runtime, Project: t.Project, Profile: t.Profile}
	s, err := snapshot(t.manifestPath(e))
	if err != nil {
		return m, err
	}
	if !s.Exists {
		return m, nil
	}
	if s.Link != "" {
		return m, fmt.Errorf("manifest is a symlink")
	}
	if err = json.Unmarshal(s.Content, &m); err != nil {
		return m, err
	}
	if m.Format != 1 || m.Scope != t.Scope || m.Runtime != t.Runtime || m.Project != t.Project || m.Profile != t.Profile {
		return m, fmt.Errorf("manifest scope/identity mismatch")
	}
	return m, nil
}
func resourceDigest(r Resource, s Snapshot) (string, bool, error) {
	if !s.Exists {
		return "", false, nil
	}
	switch r.Kind {
	case "link":
		if s.Link == "" {
			return "", true, nil
		}
		return fingerprint(s.Link), true, nil
	case "text":
		if s.Link != "" {
			return "", true, fmt.Errorf("instructions are a symlink")
		}
		content, a, _, err := ownedText(s.Content)
		return bytesDigest(content), a >= 0, err
	case "file":
		if s.Link != "" {
			return "", true, fmt.Errorf("owned file replaced with symlink")
		}
		return bytesDigest(s.Content), true, nil
	case "toml", "json":
		if s.Link != "" {
			return "", true, fmt.Errorf("configuration is a symlink")
		}
		d, err := parseDocument(s.Content, r.Kind == "toml")
		if err != nil {
			return "", false, err
		}
		v, exists := lookup(d, r.Key)
		return fingerprint(v), exists, nil
	default:
		return "", false, fmt.Errorf("unknown resource kind %s", r.Kind)
	}
}
func buildPlan(e Environment, t Target, desired []Resource, uninstall bool) (*Plan, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	old, err := loadManifest(e, t)
	if err != nil {
		return nil, err
	}
	previous := map[string]Resource{}
	for _, r := range old.Resources {
		if err := validateOwnedResource(e, t, r); err != nil {
			return nil, err
		}
		if _, ok := previous[resourceID(r)]; ok {
			return nil, fmt.Errorf("duplicate ownership")
		}
		previous[resourceID(r)] = r
	}
	want := map[string]Resource{}
	for _, r := range desired {
		if err := validateOwnedResource(e, t, r); err != nil {
			return nil, err
		}
		id := resourceID(r)
		if _, ok := want[id]; ok {
			return nil, fmt.Errorf("duplicate selected destination %s", r.Destination)
		}
		want[id] = r
	}
	snapshots := map[string]Snapshot{}
	working := map[string]Snapshot{}
	order := []string{}
	get := func(path string) (Snapshot, error) {
		if s, ok := working[path]; ok {
			return s, nil
		}
		if err := safeParents(path); err != nil {
			return Snapshot{}, err
		}
		s, err := snapshot(path)
		if err == nil {
			snapshots[path] = s
			working[path] = s
			order = append(order, path)
		}
		return s, err
	}
	p := &Plan{}
	// Validate all old identities, including entries being retained.
	for _, r := range old.Resources {
		s, err := get(r.Destination)
		if err != nil {
			return nil, err
		}
		digest, exists, err := resourceDigest(r, s)
		if err != nil {
			return nil, err
		}
		if !exists || digest != r.Digest {
			return nil, fmt.Errorf("owned resource changed or missing; refusing %s", r.Destination)
		}
	}
	for _, r := range old.Resources {
		if _, keep := want[resourceID(r)]; keep {
			continue
		}
		s := working[r.Destination]
		if r.Kind == "text" {
			s.Content, err = changeText(s.Content, nil, true)
		} else if r.Kind == "toml" || r.Kind == "json" {
			s.Content, err = changeEntry(s.Content, r, nil, true)
		} else {
			s.Exists = false
			s.Content = nil
			s.Link = ""
		}
		if err != nil {
			return nil, err
		}
		working[r.Destination] = s
		p.Actions = append(p.Actions, "REMOVE "+r.Capability+" → "+r.Destination)
	}
	for _, r := range desired {
		s, err := get(r.Destination)
		if err != nil {
			return nil, err
		}
		digest, exists, err := resourceDigest(r, s)
		if err != nil {
			return nil, err
		}
		_, owned := previous[resourceID(r)]
		if exists && !owned {
			return nil, fmt.Errorf("unowned destination already exists: %s %v", r.Destination, r.Key)
		}
		if owned && digest == r.Digest {
			continue
		}
		if r.Kind == "link" {
			if s.Exists {
				return nil, fmt.Errorf("link target changed; uninstall before replacing %s", r.Destination)
			}
			s.Link = r.Link
			s.Exists = true
		} else if r.Kind == "text" {
			s.Content, err = changeText(s.Content, r.Content, false)
			s.Exists = true
			if err != nil {
				return nil, err
			}
		} else if r.Kind == "file" {
			s.Content = r.Content
			s.Exists = true
		} else {
			s.Content, err = changeEntry(s.Content, r, r.Value, false)
			s.Exists = true
			if err != nil {
				return nil, err
			}
		}
		working[r.Destination] = s
		p.Actions = append(p.Actions, "INSTALL "+r.Capability+" → "+r.Destination)
	}
	for _, path := range order {
		before, after := snapshots[path], working[path]
		if !sameSnapshot(before, after) {
			p.Changes = append(p.Changes, Mutation{before, after})
		}
	}
	mp := t.manifestPath(e)
	before, err := snapshot(mp)
	if err != nil {
		return nil, err
	}
	after := before
	if uninstall {
		after.Exists = false
		after.Content = nil
	} else {
		old.Resources = desired
		b, err := json.MarshalIndent(old, "", "  ")
		if err != nil {
			return nil, err
		}
		after.Content = append(b, '\n')
		after.Exists = true
		after.Mode = 0600
	}
	if !sameSnapshot(before, after) {
		p.Changes = append(p.Changes, Mutation{before, after})
	}
	return p, nil
}
func changeEntry(b []byte, r Resource, v any, remove bool) ([]byte, error) {
	if r.Kind == "toml" {
		return changeTOML(b, r.Key, v, remove)
	}
	return changeJSON(b, r.Key, v, remove)
}
func makeEntry(t Target, capability, source, version, dest string, key []string, value any) Resource {
	kind := "json"
	if t.Runtime == "codex" {
		kind = "toml"
	}
	return Resource{Capability: capability, Source: source, Version: version, Scope: t.Scope, Destination: dest, Runtime: t.Runtime, Kind: kind, Key: key, Digest: fingerprint(value), Value: value}
}
func planUninstall(e Environment, t Target) (*Plan, error) { return buildPlan(e, t, nil, true) }
