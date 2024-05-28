package main

import (
	"github.com/ynzwtks/hc/cmd"
)

var version string

func main() {
	cmd.Version = version
	cmd.Execute()
}
