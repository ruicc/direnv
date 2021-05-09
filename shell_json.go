package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

// jsonShell is not a real shell
type jsonShell struct{}

// JSON is not really a shell but it fits. Useful to add support to editor and
// other external tools that understand JSON as a format.
var JSON Shell = jsonShell{}

func (sh jsonShell) Hook() (string, error) {
	return "", errors.New("this feature is not supported")
}

func (sh jsonShell) Export(e *ShellExport) string {
	out, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		// Should never happen
		panic(err)
	}
	return string(out)
}

func (sh jsonShell) Dump(env *Env) string {
	out, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		// Should never happen
		panic(err)
	}
	return string(out)
}

func (sh jsonShell) ParseAliases(rawAliases []byte) (map[string]string, error) {
	return nil, fmt.Errorf("Aliases are not supported in json")
}
