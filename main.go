package main

import (
	"fmt"
	"github.com/ynzwtks/ahccli/cmd"
)

var version string

func main() {
	fmt.Println(version)
	cmd.Execute()
}
