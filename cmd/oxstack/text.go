package main

import (
	"bytes"
	"fmt"
)

var textStart = []byte("<!-- oxStack:instructions:start -->\n")
var textEnd = []byte("<!-- oxStack:instructions:end -->\n")

func ownedText(b []byte) ([]byte, int, int, error) {
	a := bytes.Index(b, textStart)
	z := bytes.Index(b, textEnd)
	if a < 0 && z < 0 {
		return nil, -1, -1, nil
	}
	if a < 0 || z < a || bytes.Count(b, textStart) != 1 || bytes.Count(b, textEnd) != 1 {
		return nil, 0, 0, fmt.Errorf("ambiguous oxStack instruction block")
	}
	return b[a+len(textStart) : z], a, z + len(textEnd), nil
}
func changeText(b, content []byte, remove bool) ([]byte, error) {
	_, a, z, err := ownedText(b)
	if err != nil {
		return nil, err
	}
	if a >= 0 {
		b = append(append([]byte{}, b[:a]...), b[z:]...)
	}
	if remove {
		return b, nil
	}
	block := append(append(append([]byte{}, textStart...), content...), textEnd...)
	if len(b) > 0 && b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}
	return append(b, block...), nil
}
