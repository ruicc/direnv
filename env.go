package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/direnv/direnv/v2/gzenv"
)

// Env is a map representation of a shell context (support: environment variables and aliases).
type Env struct {
	EnvVars map[string]string
	Aliases map[string]string
}

func NewEnv() *Env {
	return &Env{
		EnvVars: make(map[string]string),
		Aliases: make(map[string]string),
	}
}

// GetEnv turns the classic unix environment variables into a map of
// key->values which is more handy to work with.
//
// NOTE:  We don't support having two variables with the same name.
//        I've never seen it used in the wild but accoding to POSIX
//        it's allowed.
func GetEnv() *Env {
	env := NewEnv()
	for _, kv := range os.Environ() {
		kv2 := strings.SplitN(kv, "=", 2)

		key := kv2[0]
		value := kv2[1]

		env.EnvVars[key] = value
		// TODO: Init aliases
	}

	return env
}

// CleanContext removes all the direnv-related environment variables. Call
// this after reverting the environment, otherwise direnv will just be amnesic
// about the previously-loaded environment.
func (env Env) CleanContext() {
	delete(env.EnvVars, DIRENV_DIFF)
	delete(env.EnvVars, DIRENV_DIR)
	delete(env.EnvVars, DIRENV_DUMP_FILE_PATH)
	delete(env.EnvVars, DIRENV_WATCHES)
}

// LoadEnv unmarshals the env back from a gzenv string
func LoadEnv(gzenvStr string) (env *Env, err error) {
	env = NewEnv()
	err = gzenv.Unmarshal(gzenvStr, &env)
	return
}

// LoadEnvJSON unmarshals the env back from a JSON string
func LoadEnvJSON(jsonBytes []byte) (env *Env, err error) {
	env = NewEnv()
	err = json.Unmarshal(jsonBytes, &env)
	return env, err
}

// Copy returns a fresh copy of the env. Because the env is a map under the
// hood, we want to get a copy whenever we mutate it and want to keep the
// original around.
func (env *Env) Copy() *Env {
	newEnv := NewEnv()

	for key, value := range env.EnvVars {
		newEnv.EnvVars[key] = value
	}
	for key, value := range env.Aliases {
		newEnv.Aliases[key] = value
	}

	return newEnv
}

// ToGoEnv should really be named ToUnixEnv. It turns the env back into a list
// of "key=value" strings like returns by os.Environ().
func (env Env) ToGoEnv() []string {
	goEnv := make([]string, len(env.EnvVars)) // TODO: Aliases?
	index := 0
	for key, value := range env.EnvVars {
		goEnv[index] = strings.Join([]string{key, value}, "=")
		index++
	}
	return goEnv
}

// ToShell outputs the environment into an evaluatable string that is
// understood by the target shell
func (env Env) ToShell(shell Shell) string {
	e := NewShellExport()

	for key, value := range env.EnvVars {
		e.Add(key, value)
	}

	return shell.Export(e)
}

// Serialize marshals the env into the gzenv format
func (env Env) Serialize() string {
	return gzenv.Marshal(env)
}

// Diff returns the diff between the current env and the passed env
func (env Env) Diff(other *Env) *EnvDiff {
	return BuildEnvDiff(&env, other)
}

// Fetch tries to get the value associated with the given 'key', or returns
// the provided default if none is set.
//
// Note that empty environment variables are considered to be set.
func (env Env) Fetch(key, def string) string {
	v, ok := env.EnvVars[key]
	if !ok {
		v = def
	}
	return v
}
