package main

import (
	"github.com/tzvetkoff/dockerhub-watcher/commands"
)

func main() {
	_ = commands.NewRootCommand().Execute()
}
