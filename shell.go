package main

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
)

// Shell is the interface that represents the interaction with the host shell.
type Shell interface {
	// Hook is the string that gets evaluated into the host shell config and
	// setups direnv as a prompt hook.
	Hook() (string, error)

	// Export outputs the ShellExport as an evaluatable string on the host shell
	Export(e *ShellExport) string

	// Dump outputs and evaluatable string that sets the env in the host shell
	Dump(env *Env) string

	ParseAliases(rawAliases []byte) (map[string]string, error)
}

// ShellExport represents environment variables to add and remove on the host
// shell.
type ShellExport struct {
	EnvVars map[string]*string
	Aliases map[string]*string
}

func NewShellExport() *ShellExport {
	return &ShellExport{
		EnvVars: make(map[string]*string),
		Aliases: make(map[string]*string),
	}
}

// Add represents the additon of a new environment variable
func (e ShellExport) AddEnvVar(key, value string) {
	e.EnvVars[key] = &value
}
func (e ShellExport) AddAlias(key, value string) {
	e.Aliases[key] = &value
}

// Remove represents the removal of a given `key` environment variable.
func (e ShellExport) RemoveEnvVar(key string) {
	e.EnvVars[key] = nil
}
func (e ShellExport) RemoveAlias(key string) {
	e.Aliases[key] = nil
}

// DetectShell returns a Shell instance from the given target.
//
// target is usually $0 and can also be prefixed by `-`
func DetectShell(target string) Shell {
	target = filepath.Base(target)
	// $0 starts with "-"
	if target[0:1] == "-" {
		target = target[1:]
	}

	switch target {
	case "bash":
		return Bash
	case "zsh":
		return Zsh
	case "fish":
		return Fish
	case "gzenv":
		return GzEnv
	case "vim":
		return Vim
	case "tcsh":
		return Tcsh
	case "json":
		return JSON
	case "elvish":
		return Elvish
	}

	return nil
}

func ParseAliases(rawAliases []byte, prefixLen int, separator string, enclosure string) (map[string]string, error) {
	aliasMap := make(map[string]string)
	sc := bufio.NewScanner(strings.NewReader(string(rawAliases)))
	for sc.Scan() {
		line := sc.Text()
		if len(line) <= 0 {
			continue
		}
		eqIdx := strings.Index(line, separator)
		if eqIdx == -1 {
			return nil, fmt.Errorf("'%s' not found in zsh alias line: %s", separator, line)
		}
		key := strings.Trim(line[prefixLen:eqIdx], enclosure)
		val := strings.Trim(line[eqIdx+1:], enclosure)
		aliasMap[key] = val
	}
	return aliasMap, nil
}
