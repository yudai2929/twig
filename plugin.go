package twig

import _ "embed"

//go:embed zsh/twig.zsh
var pluginScript []byte

func PluginScript() []byte {
	return pluginScript
}
