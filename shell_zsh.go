package main

// ZSH is a singleton instance of ZSH_T
type zsh struct{}

// Zsh adds support for the venerable Z shell.
var Zsh Shell = zsh{}

const zshHook = `
_direnv_hook() {
  local alias_list=$(mktemp)
  trap -- "rm -f $alias_list" SIGINT;
  alias >> "$alias_list"
  eval "$("{{.SelfPath}}" export zsh "$alias_list")";
  trap - SIGINT;
  rm -f "$alias_list"
}
typeset -ag precmd_functions;
if [[ -z ${precmd_functions[(r)_direnv_hook]} ]]; then
  precmd_functions=( _direnv_hook ${precmd_functions[@]} )
fi
typeset -ag chpwd_functions;
if [[ -z ${chpwd_functions[(r)_direnv_hook]} ]]; then
  chpwd_functions=( _direnv_hook ${chpwd_functions[@]} )
fi
`

func (sh zsh) Hook() (string, error) {
	return zshHook, nil
}

func (sh zsh) Export(e *ShellExport) (out string) {
	for key, value := range e.EnvVars {
		if value == nil {
			out += sh.unset(key)
		} else {
			out += sh.export(key, *value)
		}
	}
	for key, value := range e.Aliases {
		if value == nil {
			out += sh.unalias(key)
		} else {
			out += sh.alias(key, *value)
		}
	}
	logDebug("Export(zsh): %s", out)
	return out
}

func (sh zsh) Dump(env *Env) (out string) {
	for key, value := range env.EnvVars {
		out += sh.export(key, value)
	}
	return out
}

func (sh zsh) ParseAliases(rawAliases []byte) (map[string]string, error) {
	return ParseAliases(rawAliases, 0, "=", "'")
}

func (sh zsh) export(key, value string) string {
	return "export " + sh.escape(key) + "=" + sh.escape(value) + ";"
}

func (sh zsh) unset(key string) string {
	return "unset " + sh.escape(key) + ";"
}

func (sh zsh) escape(str string) string {
	return BashEscape(str)
}

func (sh zsh) alias(key, value string) string {
	return "alias " + sh.escape(key) + "=" + sh.escape(value) + ";"
}

func (sh zsh) unalias(key string) string {
	return "unalias " + sh.escape(key) + ";"
}
