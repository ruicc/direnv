package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
)

// CmdExport is `direnv export $0`
var CmdExport = &Cmd{
	Name:    "export",
	Desc:    "loads an .envrc and prints the diff in terms of exports",
	Args:    []string{"SHELL", "shell_context_path"},
	Private: true,
	Action:  cmdWithWarnTimeout(actionWithConfig(exportCommand)),
}

func exportCommand(currentEnv *Env, args []string, config *Config) (err error) {
	defer log.SetPrefix(log.Prefix())
	log.SetPrefix(log.Prefix() + "export:")
	logDebug("start")

	var target string

	if len(args) > 1 {
		target = args[1]
	}

	shell := DetectShell(target)
	if shell == nil {
		return fmt.Errorf("unknown target shell '%s'", target)
	}

	if len(args) > 2 && config.EnableAliasExport {
		aliasListPath := args[2]
		var rawAliases []byte
		rawAliases, err = ioutil.ReadFile(aliasListPath)
		if err != nil {
			err = fmt.Errorf("Reading alias list failed: %w", err)
			logDebug("err: %v", err)
			return
		}
		var aliasMap map[string]string
		aliasMap, err = ParseAliases(rawAliases, 0, "=", "'") // TODO: Define shell.ParseAliases(rawAliases)
		if err != nil {
			err = fmt.Errorf("Parsing alias list failed: %w", err)
			logDebug("err: %v", err)
			return
		}
		logDebug("aliasMap: %v", aliasMap)
		currentEnv.Aliases = aliasMap
	}

	logDebug("loading RCs")
	loadedRC := config.LoadedRC()
	toLoad := findUp(config.WorkDir, ".envrc")

	if loadedRC == nil && toLoad == "" {
		return
	}

	logDebug("updating RC")
	log.SetPrefix(log.Prefix() + "update:")

	logDebug("Determining action:")
	logDebug("toLoad: %#v", toLoad)
	logDebug("loadedRC: %#v", loadedRC)

	switch {
	case toLoad == "":
		logDebug("no RC found, unloading")
	case loadedRC == nil:
		logDebug("no RC (implies no DIRENV_DIFF),loading")
	case loadedRC.path != toLoad:
		logDebug("new RC, loading")
	case loadedRC.times.Check() != nil:
		logDebug("file changed, reloading")
	default:
		logDebug("no update needed")
		return
	}

	var previousEnv, newEnv *Env

	if previousEnv, err = config.Revert(currentEnv); err != nil {
		err = fmt.Errorf("Revert() failed: %w", err)
		logDebug("err: %v", err)
		return
	}

	if toLoad == "" {
		logStatus(currentEnv, "unloading")
		newEnv = previousEnv.Copy()
		newEnv.CleanContext()
	} else {
		newEnv, err = config.EnvFromRC(toLoad, previousEnv)
		if err != nil {
			logDebug("err: %v", err)
			// If loading fails, fall through and deliver a diff anyway,
			// but still exit with an error.  This prevents retrying on
			// every prompt.
		}
		if newEnv == nil {
			// unless of course, the error was in hashing and timestamp loading,
			// in which case we have to abort because we don't know what timestamp
			// to put in the diff!
			return
		}
	}

	envvarStat, aliasStat := diffStatus(previousEnv.Diff(newEnv))
	if envvarStat != "" {
		logStatus(currentEnv, "export %s", envvarStat)
	}
	if aliasStat != "" {
		logStatus(currentEnv, "alias %s", aliasStat)
	}

	diffString := currentEnv.Diff(newEnv).ToShell(shell)
	logDebug("env diff %s", diffString)
	fmt.Print(diffString)
	return
}

// Return a string of +/-/~ indicators of an environment diff
func diffStatus(oldDiff *EnvDiff) (string, string) {
	if oldDiff.Any() {
		var envvarOut []string
		var aliasOut []string
		for key := range oldDiff.PrevEnvVars {
			_, ok := oldDiff.NextEnvVars[key]
			if !ok && !direnvKey(key) {
				envvarOut = append(envvarOut, "-"+key)
			}
		}

		for key := range oldDiff.NextEnvVars {
			_, ok := oldDiff.PrevEnvVars[key]
			if direnvKey(key) {
				continue
			}
			if ok {
				envvarOut = append(envvarOut, "~"+key)
			} else {
				envvarOut = append(envvarOut, "+"+key)
			}
		}

		for key := range oldDiff.PrevAliases {
			_, ok := oldDiff.NextAliases[key]
			if !ok {
				aliasOut = append(aliasOut, "-"+key)
			}
		}

		for key := range oldDiff.NextAliases {
			_, ok := oldDiff.PrevAliases[key]
			if ok {
				aliasOut = append(aliasOut, "~"+key)
			} else {
				aliasOut = append(aliasOut, "+"+key)
			}
		}

		sort.Strings(envvarOut)
		sort.Strings(aliasOut)
		return strings.Join(envvarOut, " "), strings.Join(aliasOut, " ")
	}
	return "", ""
}

func direnvKey(key string) bool {
	return strings.HasPrefix(key, "DIRENV_")
}
