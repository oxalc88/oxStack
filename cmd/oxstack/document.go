package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

func readDocument(path string, isTOML bool) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	return parseDocument(b, isTOML)
}
func parseDocument(b []byte, isTOML bool) (map[string]any, error) {
	d := map[string]any{}
	if len(bytes.TrimSpace(b)) == 0 {
		return d, nil
	}
	var err error
	if isTOML {
		err = toml.Unmarshal(b, &d)
	} else {
		decoder := json.NewDecoder(bytes.NewReader(b))
		decoder.UseNumber()
		err = decoder.Decode(&d)
		if err == nil {
			var extra any
			if trailing := decoder.Decode(&extra); trailing != io.EOF {
				err = fmt.Errorf("configuration has invalid trailing JSON")
			}
		}
	}
	if d == nil && err == nil {
		err = fmt.Errorf("configuration must be an object")
	}
	return d, err
}
func lookup(d map[string]any, path []string) (any, bool) {
	var v any = d
	for _, k := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok = m[k]
		if !ok {
			return nil, false
		}
	}
	return v, true
}
func changeJSON(b []byte, path []string, value any, remove bool) ([]byte, error) {
	d, err := parseDocument(b, false)
	if err != nil {
		return nil, err
	}
	m := d
	for _, k := range path[:len(path)-1] {
		v, exists := m[k]
		if !exists {
			v = map[string]any{}
			m[k] = v
		}
		var ok bool
		m, ok = v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("configuration parent %s is not an object", k)
		}
	}
	k := path[len(path)-1]
	if remove {
		delete(m, k)
	} else {
		m[k] = value
	}
	out, err := json.MarshalIndent(d, "", "  ")
	return append(out, '\n'), err
}
func keyParts(n *unstable.Node) []string {
	var k []string
	it := n.Key()
	for it.Next() {
		k = append(k, string(it.Node().Data))
	}
	return k
}
func prefix(k, p []string) bool { return len(k) >= len(p) && reflect.DeepEqual(k[:len(p)], p) }

type textRange struct{ start, end int }

// Remove expressions for exactly one owned subtree. The parser handles quoted
// keys, multiline strings and nested tables; other bytes remain untouched.
// Shared inline ancestors are deliberately refused instead of reserialized.
func stripTOML(b []byte, path []string) ([]byte, error) {
	var parser unstable.Parser
	parser.Reset(b)
	var table []string
	var cuts []textRange
	for parser.NextExpression() {
		n := parser.Expression()
		if n.Kind != unstable.Table && n.Kind != unstable.ArrayTable && n.Kind != unstable.KeyValue {
			continue
		}
		k := keyParts(n)
		if n.Kind == unstable.Table || n.Kind == unstable.ArrayTable {
			table = k
			if prefix(k, path) {
				first := n.Child()
				start := bytes.LastIndexByte(b[:int(first.Raw.Offset)], '\n') + 1
				end := start
				if x := bytes.IndexByte(b[start:], '\n'); x >= 0 {
					end = start + x + 1
				} else {
					end = len(b)
				}
				cuts = append(cuts, textRange{start, end})
			}
			continue
		}
		full := append(append([]string{}, table...), k...)
		if prefix(full, path) {
			cuts = append(cuts, textRange{int(n.Raw.Offset), int(n.Raw.Offset + n.Raw.Length)})
		} else if prefix(path, full) {
			return nil, fmt.Errorf("cannot safely edit shared inline ancestor %s", strings.Join(full, "."))
		}
	}
	if err := parser.Error(); err != nil {
		return nil, err
	}
	var out []byte
	pos := 0
	for _, r := range cuts {
		out = append(out, b[pos:r.start]...)
		pos = r.end
	}
	out = append(out, b[pos:]...)
	return out, nil
}
func changeTOML(b []byte, path []string, value any, remove bool) ([]byte, error) {
	d, err := parseDocument(b, true)
	if err != nil {
		return nil, err
	}
	if _, exists := lookup(d, path); exists {
		b, err = stripTOML(b, path)
		if err != nil {
			return nil, err
		}
	}
	if !remove {
		nested := value
		for i := len(path) - 1; i >= 0; i-- {
			nested = map[string]any{path[i]: nested}
		}
		encoded, err := toml.Marshal(nested)
		if err != nil {
			return nil, err
		}
		// Drop encoder-generated ancestor headers; only owned headers are appended.
		var parser unstable.Parser
		parser.Reset(encoded)
		var cuts []textRange
		for parser.NextExpression() {
			n := parser.Expression()
			if n.Kind == unstable.Table {
				k := keyParts(n)
				if len(k) < len(path) && prefix(path, k) {
					start := bytes.LastIndexByte(encoded[:int(n.Child().Raw.Offset)], '\n') + 1
					end := start + bytes.IndexByte(encoded[start:], '\n') + 1
					cuts = append(cuts, textRange{start, end})
				}
			}
		}
		pos := 0
		var owned []byte
		for _, r := range cuts {
			owned = append(owned, encoded[pos:r.start]...)
			pos = r.end
		}
		owned = append(owned, encoded[pos:]...)
		if len(path) == 1 {
			b = append(append(owned, '\n'), b...)
		} else {
			b = append(append(b, '\n'), owned...)
		}
	}
	if _, err := parseDocument(b, true); err != nil {
		return nil, fmt.Errorf("unsafe TOML edit refused: %w", err)
	}
	return b, nil
}
