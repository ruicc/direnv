package main

import (
	"strings"

	"github.com/direnv/direnv/v2/gzenv"
)

// IgnoredKeys is list of keys we don't want to deal with
var IgnoredKeys = map[string]bool{
	// direnv env config
	"DIRENV_CONFIG": true,
	"DIRENV_BASH":   true,

	// should only be available inside of the .envrc
	"DIRENV_IN_ENVRC": true,

	"COMP_WORDBREAKS": true, // Avoids segfaults in bash
	"PS1":             true, // PS1 should not be exported, fixes problem in bash

	// variables that should change freely
	"OLDPWD":    true,
	"PWD":       true,
	"SHELL":     true,
	"SHELLOPTS": true,
	"SHLVL":     true,
	"_":         true,
}
var IgnoredAliasKeys = map[string]bool{
	// Special alias
	"-": true,
}

// EnvDiff represents the diff between two environments
type EnvDiff struct {
	PrevEnvVars map[string]string `json:"pe"`
	NextEnvVars map[string]string `json:"ne"`
	PrevAliases map[string]string `json:"pa"`
	NextAliases map[string]string `json:"na"`
}

// NewEnvDiff is an empty constructor for EnvDiff
func NewEnvDiff() *EnvDiff {
	return &EnvDiff{
		make(map[string]string),
		make(map[string]string),
		make(map[string]string),
		make(map[string]string),
	}
}

// BuildEnvDiff analyses the changes between 'e1' and 'e2' and builds an
// EnvDiff out of it.
func BuildEnvDiff(e1, e2 *Env) *EnvDiff {
	diff := NewEnvDiff()

	in := func(key string, m map[string]string) bool {
		_, ok := m[key]
		return ok
	}

	// Handle EnvVars
	for key := range e1.EnvVars {
		if IgnoredEnv(key) {
			continue
		}
		if e2.EnvVars[key] != e1.EnvVars[key] || !in(key, e2.EnvVars) {
			diff.PrevEnvVars[key] = e1.EnvVars[key]
		}
	}

	for key := range e2.EnvVars {
		if IgnoredEnv(key) {
			continue
		}
		if e2.EnvVars[key] != e1.EnvVars[key] || !in(key, e1.EnvVars) {
			diff.NextEnvVars[key] = e2.EnvVars[key]
		}
	}

	// Handle Aliases
	for key := range e1.Aliases {
		if IgnoredAlias(key) {
			continue
		}
		if e2.Aliases[key] != e1.Aliases[key] || !in(key, e2.Aliases) {
			diff.PrevAliases[key] = e1.Aliases[key]
		}
	}

	for key := range e2.Aliases {
		if IgnoredAlias(key) {
			continue
		}
		if e2.Aliases[key] != e1.Aliases[key] || !in(key, e1.Aliases) {
			diff.NextAliases[key] = e2.Aliases[key]
		}
	}

	return diff
}

// LoadEnvDiff unmarshalls a gzenv string back into an EnvDiff.
func LoadEnvDiff(gzenvStr string) (diff *EnvDiff, err error) {
	diff = new(EnvDiff)
	err = gzenv.Unmarshal(gzenvStr, diff)
	return
}

// Any returns if the diff contains any changes.
func (diff *EnvDiff) Any() bool {
	return len(diff.PrevEnvVars) > 0 || len(diff.NextEnvVars) > 0 ||
		len(diff.PrevAliases) > 0 || len(diff.NextAliases) > 0
}

// ToShell applies the env diff as a set of commands that are understood by
// the target `shell`. The outputted string is then meant to be evaluated in
// the target shell.
func (diff *EnvDiff) ToShell(shell Shell) string {
	e := NewShellExport()

	for key := range diff.PrevEnvVars {
		_, ok := diff.NextEnvVars[key]
		if !ok {
			e.RemoveEnvVar(key)
		}
	}

	for key, value := range diff.NextEnvVars {
		e.AddEnvVar(key, value)
	}

	for key := range diff.PrevAliases {
		_, ok := diff.NextAliases[key]
		if !ok {
			e.RemoveAlias(key)
		}
	}

	for key, value := range diff.NextAliases {
		e.AddAlias(key, value)
	}

	return shell.Export(e)
}

// Patch applies the diff to the given env and returns a new env with the
// changes applied.
func (diff *EnvDiff) Patch(env *Env) (newEnv *Env) {
	newEnv = NewEnv()

	for k, v := range env.EnvVars {
		newEnv.EnvVars[k] = v
	}

	for key := range diff.PrevEnvVars {
		delete(newEnv.EnvVars, key)
	}

	for key, value := range diff.NextEnvVars {
		newEnv.EnvVars[key] = value
	}

	for k, v := range env.Aliases {
		newEnv.Aliases[k] = v
	}

	for key := range diff.PrevAliases {
		delete(newEnv.Aliases, key)
	}

	for key, value := range diff.NextAliases {
		newEnv.Aliases[key] = value
	}

	return newEnv
}

// Reverse flips the diff so that it applies the other way around.
func (diff *EnvDiff) Reverse() *EnvDiff {
	return &EnvDiff{
		diff.NextEnvVars,
		diff.PrevEnvVars,
		diff.NextAliases,
		diff.PrevAliases,
	}
}

// Serialize marshalls the environment diff to the gzenv format.
func (diff *EnvDiff) Serialize() string {
	return gzenv.Marshal(diff)
}

//// Utils

// IgnoredEnv returns true if the key should be ignored in environment diffs.
func IgnoredEnv(key string) bool {
	if strings.HasPrefix(key, "__fish") {
		return true
	}
	if strings.HasPrefix(key, "BASH_FUNC_") {
		return true
	}
	_, found := IgnoredKeys[key]
	return found
}

func IgnoredAlias(key string) bool {
	_, found := IgnoredAliasKeys[key]
	return found
}
